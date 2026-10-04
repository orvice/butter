import { createContext, useContext } from 'react'

// ChatAgent is the agent that answers in a thread: each reply shows its
// avatar and name.
export interface ChatAgent {
  name: string
  iconUrl?: string
}

export const ChatAgentContext = createContext<ChatAgent>({ name: 'Agent' })

export function useChatAgent(): ChatAgent {
  return useContext(ChatAgentContext)
}
