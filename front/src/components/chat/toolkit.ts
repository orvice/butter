import { AuiConfig, Tools, defineToolkit } from '@assistant-ui/react'
import { HumanInputToolView, RenderUIToolView } from './tool-views'

// chatToolkit names the tools a chat draws its own way. Butter runs them on
// the server, so they are `backend` entries: they only render, and are never
// offered to the agent as tools the browser runs. Every other tool call is a
// ToolCallView, the `tools.Fallback` of a message's parts.
export const chatToolkit = defineToolkit({
  // render_ui draws an A2UI card, which the reply shows instead of the call.
  render_ui: { type: 'backend', render: RenderUIToolView },
  // adk_request_input carries a Workflow's Human Input question. Over AG-UI
  // the question arrives as an Interrupt instead, never as this call.
  adk_request_input: { type: 'backend', render: HumanInputToolView },
})

// chatAuiConfig installs the toolkit, as the `config` of a chat's
// AssistantRuntimeProvider.
export const chatAuiConfig = AuiConfig({
  tools: Tools({ toolkit: chatToolkit }),
})
