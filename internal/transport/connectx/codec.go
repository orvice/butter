package connectx

import (
	"bytes"
	"encoding/json"
	"fmt"

	"connectrpc.com/connect"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

// snakeCaseJSONCodec replaces Connect's default "json" codec with one that
// emits proto field names (snake_case) instead of JSON names (camelCase) on
// the wire. The previous Twirp transport defaulted to snake_case JSON, and
// 130+ dashboard call sites read response fields like `connected_daemons` /
// `base_url` / `space_id` without a camelCase fallback. Until those callers
// are migrated to typed Connect clients (which use camelCase natively), the
// server must keep speaking snake_case so the dashboard stays functional.
//
// The unmarshal direction sets DiscardUnknown=true to remain forward
// compatible with clients that send additional fields, mirroring Connect's
// default behavior.
type snakeCaseJSONCodec struct {
	name string
}

func newSnakeCaseJSONCodec(name string) *snakeCaseJSONCodec {
	return &snakeCaseJSONCodec{name: name}
}

func (c *snakeCaseJSONCodec) Name() string { return c.name }

func (c *snakeCaseJSONCodec) Marshal(message any) ([]byte, error) {
	msg, ok := message.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("connectx: marshal: expected proto.Message, got %T", message)
	}
	return protojson.MarshalOptions{UseProtoNames: true}.Marshal(msg)
}

func (c *snakeCaseJSONCodec) MarshalAppend(dst []byte, message any) ([]byte, error) {
	msg, ok := message.(proto.Message)
	if !ok {
		return nil, fmt.Errorf("connectx: marshal: expected proto.Message, got %T", message)
	}
	return protojson.MarshalOptions{UseProtoNames: true}.MarshalAppend(dst, msg)
}

func (c *snakeCaseJSONCodec) MarshalStable(message any) ([]byte, error) {
	raw, err := c.Marshal(message)
	if err != nil {
		return nil, err
	}
	// protojson emits non-deterministic whitespace; compact it so stable
	// callers get a byte-stable representation. Matches the upstream impl.
	out := &bytes.Buffer{}
	if err := json.Compact(out, raw); err != nil {
		return nil, err
	}
	return out.Bytes(), nil
}

func (c *snakeCaseJSONCodec) Unmarshal(data []byte, message any) error {
	msg, ok := message.(proto.Message)
	if !ok {
		return fmt.Errorf("connectx: unmarshal: expected proto.Message, got %T", message)
	}
	if len(data) == 0 {
		// protojson rejects empty input; Connect peers send "{}" for empty
		// messages but be lenient with literal empty bodies too.
		return nil
	}
	return protojson.UnmarshalOptions{DiscardUnknown: true}.Unmarshal(data, msg)
}

func (c *snakeCaseJSONCodec) IsBinary() bool { return false }

// HandlerOptions returns the connect.HandlerOption slice that should be
// applied to every Connect handler so server responses keep their legacy
// snake_case JSON shape. The same codec is registered under both "json" and
// "json; charset=utf-8" names because connect-go matches Content-Type
// suffixes exactly.
func HandlerOptions() []connect.HandlerOption {
	return []connect.HandlerOption{
		connect.WithCodec(newSnakeCaseJSONCodec("json")),
		connect.WithCodec(newSnakeCaseJSONCodec("json; charset=utf-8")),
	}
}
