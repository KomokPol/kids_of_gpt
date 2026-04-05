import { create } from 'zustand'
import { getMoshennikResponse } from '../api/moshennik.js'

/**
 * useMoshennikStore — стор чата с MoshennikAI
 */
const useMoshennikStore = create((set, get) => ({
  // ── State ──────────────────────────────────────────────────────────────────
  messages: [
    {
      id: 1,
      role: 'ai',
      text: 'Здорово, братан. Я MoshennikAI — твой личный наставник по жизни на зоне. Задавай вопросы, учись по понятиям.',
      ts: Date.now(),
    },
  ],
  typing:      false,   // AI "печатает"
  activeCourse: null,   // выбранный курс

  // ── Actions ────────────────────────────────────────────────────────────────
  sendMessage: async (text) => {
    if (!text.trim() || get().typing) return

    const userMsg = {
      id: Date.now(),
      role: 'user',
      text: text.trim(),
      ts: Date.now(),
    }

    set(s => ({ messages: [...s.messages, userMsg], typing: true }))

    try {
      const response = await getMoshennikResponse(text)
      const aiMsg = {
        id: Date.now() + 1,
        role: 'ai',
        text: response,
        ts: Date.now(),
      }
      set(s => ({ messages: [...s.messages, aiMsg], typing: false }))
    } catch {
      set(s => ({
        messages: [...s.messages, {
          id: Date.now() + 1,
          role: 'ai',
          text: 'Связь прервалась. Попробуй ещё раз.',
          ts: Date.now(),
        }],
        typing: false,
      }))
    }
  },

  setActiveCourse: (course) => set({ activeCourse: course }),

  clearChat: () => set({
    messages: [{
      id: Date.now(),
      role: 'ai',
      text: 'Новый разговор. Спрашивай.',
      ts: Date.now(),
    }],
  }),
}))

export default useMoshennikStore
