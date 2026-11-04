package mongostore_test

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"sync"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
	ws "go.orx.me/apps/butter/internal/workstation"
	"go.orx.me/apps/butter/internal/workstation/mongostore"
	"go.orx.me/apps/butter/internal/workstation/repotest"
	v1 "go.orx.me/apps/butter/pkg/proto/agents/v1"
)

const processPrefix = "WORK_STATION_CHILD_RESULT:"

type processRequest struct {
	Operation string
	Attempt   ws.Attempt
	Title     string
	BudgetKey string
	Duration  time.Duration
}
type processResponse struct {
	Code      string
	MongoCode int
	Attempt   ws.Attempt
	Status    string
	Due       time.Time
	Wait      time.Duration
}

// This helper is a separate OS process using its own Mongo client and Service.
// Its protocol only contains test fixture identities, never real credentials.
func TestWorkStationMongoProcessHelper(t *testing.T) {
	if os.Getenv("BUTTER_WS_PROCESS_HELPER") != "1" {
		t.Skip("subprocess helper")
	}
	client, err := mongo.Connect(options.Client().ApplyURI(os.Getenv("BUTTER_TEST_MONGO_URI")).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatal("subprocess Mongo connection failed")
	}
	defer client.Disconnect(context.Background())
	repo := mongostore.New(client.Database(os.Getenv("BUTTER_WS_PROCESS_DB")))
	service := &ws.Service{Repo: repo}
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 4096), 1<<20)
	for scanner.Scan() {
		var req processRequest
		if json.Unmarshal(scanner.Bytes(), &req) != nil {
			t.Fatal("invalid subprocess request")
		}
		result := processResponse{Code: "OK"}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		switch req.Operation {
		case "claim":
			result.Attempt, err = service.Claim(ctx, "workspace-A", "connection-1", "42")
		case "publish":
			var now time.Time
			now, err = repo.Now(ctx)
			if err == nil {
				err = service.Publish(ctx, req.Attempt, ws.Scan{Coverage: "COMPLETE", Items: []*v1.WorkStationItem{repotest.Item("1", req.Title, now)}})
			}
		case "fail_auth":
			err = service.Fail(ctx, req.Attempt, &ws.Failure{Category: "AUTH_INVALID"})
		case "refresh":
			result.Status, result.Due, err = service.Refresh(ctx, "workspace-A", "connection-1")
		case "budget_reserve":
			result.Wait, err = repo.ReserveBudget(ctx, req.BudgetKey, req.Duration)
		case "budget_defer":
			err = repo.DeferBudget(ctx, req.BudgetKey, req.Duration)
		default:
			t.Fatal("unknown subprocess operation")
		}
		cancel()
		if err != nil {
			var command mongo.CommandError
			var write mongo.WriteException
			if errors.As(err, &command) {
				result.MongoCode = int(command.Code)
			}
			if errors.As(err, &write) && len(write.WriteErrors) > 0 {
				result.MongoCode = write.WriteErrors[0].Code
			}
			switch {
			case errors.Is(err, ws.ErrFenced):
				result.Code = "FENCED"
			case errors.Is(err, ws.ErrPaused):
				result.Code = "PAUSED"
			case errors.Is(err, ws.ErrNotFound):
				result.Code = "NOT_FOUND"
			default:
				result.Code = ws.Classify(err).Category
			}
		}
		encoded, _ := json.Marshal(result)
		fmt.Println(processPrefix + string(encoded))
	}
	if scanner.Err() != nil {
		t.Fatal("subprocess request stream failed")
	}
}

type mongoProcess struct {
	cmd     *exec.Cmd
	input   io.WriteCloser
	output  *bufio.Scanner
	stderr  bytes.Buffer
	cancel  context.CancelFunc
	stopped bool
}

func startMongoProcess(t *testing.T, database string) *mongoProcess {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	p := &mongoProcess{cancel: cancel}
	p.cmd = exec.CommandContext(ctx, os.Args[0], "-test.run=^TestWorkStationMongoProcessHelper$")
	p.cmd.Env = append(os.Environ(), "BUTTER_WS_PROCESS_HELPER=1", "BUTTER_WS_PROCESS_DB="+database)
	p.cmd.Stderr = &p.stderr
	var err error
	p.input, err = p.cmd.StdinPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	output, err := p.cmd.StdoutPipe()
	if err != nil {
		cancel()
		t.Fatal(err)
	}
	p.output = bufio.NewScanner(output)
	p.output.Buffer(make([]byte, 4096), 1<<20)
	if err := p.cmd.Start(); err != nil {
		cancel()
		t.Fatal(err)
	}
	t.Cleanup(func() { p.stop(t, false) })
	return p
}
func (p *mongoProcess) call(req processRequest) (processResponse, error) {
	encoded, err := json.Marshal(req)
	if err != nil {
		return processResponse{}, err
	}
	if _, err := fmt.Fprintln(p.input, string(encoded)); err != nil {
		return processResponse{}, err
	}
	for p.output.Scan() {
		line := p.output.Text()
		if !strings.HasPrefix(line, processPrefix) {
			continue
		}
		var result processResponse
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, processPrefix)), &result); err != nil {
			return processResponse{}, err
		}
		return result, nil
	}
	return processResponse{}, errors.New("Mongo subprocess ended without a response")
}
func (p *mongoProcess) mustCall(t *testing.T, req processRequest) processResponse {
	t.Helper()
	result, err := p.call(req)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
func (p *mongoProcess) stop(t *testing.T, crash bool) {
	t.Helper()
	if p.stopped {
		return
	}
	p.stopped = true
	_ = p.input.Close()
	if crash {
		_ = p.cmd.Process.Kill()
	}
	err := p.cmd.Wait()
	p.cancel()
	if err != nil && !crash {
		t.Errorf("Mongo subprocess failed (stderr omitted to avoid accidental connection details)")
	}
}

func TestMongoIndependentProcessesTakeoverRefreshAndRestart(t *testing.T) {
	uri := os.Getenv("BUTTER_TEST_MONGO_URI")
	if uri == "" {
		t.Skip("BUTTER_TEST_MONGO_URI not set; real Mongo multiprocess acceptance not run")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 40*time.Second)
	defer cancel()
	client, err := mongo.Connect(options.Client().ApplyURI(uri).SetTimeout(10 * time.Second))
	if err != nil {
		t.Fatal("Mongo connection failed")
	}
	defer client.Disconnect(context.Background())
	db := client.Database(fmt.Sprintf("butter_work_station_process_test_%d", time.Now().UnixNano()))
	defer db.Drop(context.Background())
	repo := mongostore.New(db)
	if err := repo.EnsureIndexes(ctx); err != nil {
		t.Fatal(err)
	}
	now, err := repo.Now(ctx)
	if err != nil {
		t.Fatal(err)
	}
	record := ws.Record{WorkspaceID: "workspace-A", ID: "connection-1", Name: "process fixture", Source: ws.GitHubSource(), HostID: "host", PrincipalID: "fixture-user", Revision: 1, Version: 1, Epoch: "epoch-1", Credential: ws.Credential{Ciphertext: "nonsecret-fixture-placeholder"}, Scopes: []ws.Scope{{ID: "42", Name: "fixture/repo", Generation: "generation-1", Health: "QUEUED", Due: now.Add(-time.Second)}}}
	if err := repo.Create(ctx, record); err != nil {
		t.Fatal(err)
	}
	first, second := startMongoProcess(t, db.Name()), startMongoProcess(t, db.Name())
	if first.cmd.Process.Pid == second.cmd.Process.Pid {
		t.Fatal("expected independent OS processes")
	}
	responses := make([]processResponse, 2)
	errorsByProcess := make([]error, 2)
	start := make(chan struct{})
	var wg sync.WaitGroup
	for index, process := range []*mongoProcess{first, second} {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			responses[index], errorsByProcess[index] = process.call(processRequest{Operation: "claim"})
		}()
	}
	close(start)
	wg.Wait()
	for _, err := range errorsByProcess {
		if err != nil {
			t.Fatal(err)
		}
	}
	winner, loser := first, second
	claimed := responses[0]
	if claimed.Code != "OK" {
		winner, loser = second, first
		claimed = responses[1]
	}
	if claimed.Code != "OK" || responses[0].Code == responses[1].Code || (responses[0].Code != "FENCED" && responses[1].Code != "FENCED") {
		t.Fatalf("claim results=%s/%s", responses[0].Code, responses[1].Code)
	}

	// Deterministically model lease expiry in the shared database, without
	// sleeping through the production two-minute TTL or faking a process clock.
	// Publication itself still uses the real Mongo $$NOW fence.
	if _, err := db.Collection("work_station_connections").UpdateOne(ctx, bson.M{"workspace_id": record.WorkspaceID, "id": record.ID, "scopes.claim": claimed.Attempt.Token}, bson.M{"$set": bson.M{"scopes.0.lease_until": now.Add(-time.Second)}}); err != nil {
		t.Fatal(err)
	}
	takeover := loser.mustCall(t, processRequest{Operation: "claim"})
	if takeover.Code != "OK" || takeover.Attempt.Token == claimed.Attempt.Token {
		t.Fatalf("takeover=%s", takeover.Code)
	}
	if result := loser.mustCall(t, processRequest{Operation: "publish", Attempt: takeover.Attempt, Title: "takeover committed body"}); result.Code != "OK" {
		t.Fatalf("takeover publication=%s", result.Code)
	}
	for _, operation := range []string{"publish", "fail_auth"} {
		result := winner.mustCall(t, processRequest{Operation: operation, Attempt: claimed.Attempt, Title: "late private body"})
		if result.Code != "FENCED" {
			t.Fatalf("late %s result=%s", operation, result.Code)
		}
	}
	rows, err := repo.Query(ctx, record.WorkspaceID, ws.Query{Now: now, Status: "all"})
	if err != nil || len(rows) != 1 || rows[0].Title != "takeover committed body" {
		t.Fatalf("late source leaked or hid new snapshot, count=%d err=%v", len(rows), err)
	}
	current, err := repo.Get(ctx, record.WorkspaceID, record.ID)
	if err != nil || current.Paused || current.Scopes[0].Health != "HEALTHY" {
		t.Fatal("late failure changed active state", err)
	}

	refresh := winner.mustCall(t, processRequest{Operation: "refresh"})
	if refresh.Code != "OK" || refresh.Status != "QUEUED" {
		t.Fatalf("refresh=%s/%s", refresh.Code, refresh.Status)
	}
	coalesced := loser.mustCall(t, processRequest{Operation: "refresh"})
	if coalesced.Code != "OK" || coalesced.Status != "DEFERRED" || !coalesced.Due.After(now) {
		t.Fatalf("cross-process cooldown=%+v", coalesced)
	}
	// The process that accepted Refresh is killed before consuming the intent.
	// A fresh OS process must observe and consume its persisted due state.
	winner.stop(t, true)
	restarted := startMongoProcess(t, db.Name())
	recovered := restarted.mustCall(t, processRequest{Operation: "claim"})
	if recovered.Code != "OK" {
		t.Fatalf("restart lost pending intent=%s", recovered.Code)
	}
	if result := restarted.mustCall(t, processRequest{Operation: "publish", Attempt: recovered.Attempt, Title: "restart recovered body"}); result.Code != "OK" {
		t.Fatalf("restart publication=%s", result.Code)
	}
	rows, err = repo.Query(ctx, record.WorkspaceID, ws.Query{Now: time.Now(), Status: "all"})
	if err != nil || len(rows) != 1 || rows[0].Title != "restart recovered body" {
		t.Fatal("restart projection incorrect", err)
	}

	// Explicitly exercise the server-time predicate independently of a stale
	// version/token mismatch. A current version with an expired fence cannot win.
	current, err = repo.Get(ctx, record.WorkspaceID, record.ID)
	if err != nil {
		t.Fatal(err)
	}
	forged := current
	forged.Version++
	forged.Paused = true
	ok, err := repo.CAS(ctx, current, &forged, &ws.Fence{Token: current.Scopes[0].Claim, Until: now.Add(-time.Second)})
	if err != nil || ok {
		t.Fatal("expired Mongo database-time fence was accepted", err)
	}

	instanceKey, accountKey := ws.BudgetKeys(record)
	for _, key := range []string{instanceKey, accountKey} {
		start := make(chan struct{})
		responses := make([]processResponse, 2)
		errs := make([]error, 2)
		var wg sync.WaitGroup
		for index, process := range []*mongoProcess{loser, restarted} {
			wg.Add(1)
			go func() {
				defer wg.Done()
				<-start
				responses[index], errs[index] = process.call(processRequest{Operation: "budget_reserve", BudgetKey: key, Duration: 30 * time.Second})
			}()
		}
		close(start)
		wg.Wait()
		admitted := 0
		for i, result := range responses {
			if errs[i] != nil || result.Code != "OK" {
				t.Fatalf("shared reservation failed code=%s mongo_code=%d err=%v", result.Code, result.MongoCode, errs[i])
			}
			if result.Wait == 0 {
				admitted++
			}
		}
		if admitted != 1 {
			t.Fatalf("budget admitted %d independent processes", admitted)
		}
		if result := loser.mustCall(t, processRequest{Operation: "budget_defer", BudgetKey: key, Duration: 2 * time.Minute}); result.Code != "OK" {
			t.Fatalf("shared backoff=%s", result.Code)
		}
	}
	restarted.stop(t, true)
	budgetRestart := startMongoProcess(t, db.Name())
	for _, key := range []string{instanceKey, accountKey} {
		result := budgetRestart.mustCall(t, processRequest{Operation: "budget_reserve", BudgetKey: key, Duration: time.Millisecond})
		if result.Code != "OK" || result.Wait < time.Minute {
			t.Fatalf("restart bypassed shared backoff: code=%s wait=%v", result.Code, result.Wait)
		}
	}
}
