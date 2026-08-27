// Package cursorbox bridges Cursor SDK Bridge sessions hosted on a ButterBox
// into the ADK agent interface so an AGENT_TYPE_CURSOR leaf can be invoked
// like any other agent. Mirrors internal/runtime/pibox for PI agents.
//
// The bridge drives a synchronous turn API: SendMessage blocks until the
// Cursor agent finishes (the box's CursorService collects RunUpdate events
// internally), so there is no poll loop. One Cursor session exists per
// (butter session × agent), keyed in ADK session state; on a repointed agent
// or a session the box no longer knows, the bridge abandons and recreates.
//
// The box wire contract (butterbox.cursor.v1) lives in this repo's proto tree
// and mirrors butter-box issue #315; when that module publishes cursor.v1,
// this package should switch to its generated client (ADR-0011 style).
package cursorbox
