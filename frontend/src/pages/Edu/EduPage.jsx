import { useEffect, useRef, useState } from 'react'
import { Text, Button, Card, Chip } from '../../ds/index.js'
import useMoshennikStore from '../../store/useMoshennikStore.js'
import { COURSES } from '../../api/moshennik.js'
import styles from './EduPage.module.css'

export default function EduPage() {
  const { messages, typing, sendMessage, activeCourse, setActiveCourse, clearChat } = useMoshennikStore()
  const [input, setInput] = useState('')
  const messagesEndRef = useRef(null)
  const inputRef = useRef(null)

  // Автоскролл к последнему сообщению
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages, typing])

  const handleSend = () => {
    if (!input.trim() || typing) return
    sendMessage(input)
    setInput('')
    inputRef.current?.focus()
  }

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  const handleCourseClick = (course) => {
    setActiveCourse(course)
    sendMessage(`Расскажи про курс "${course.title}": ${course.description}`)
  }

  return (
    <div className={styles.page}>
      {/* Left — courses */}
      <div className={styles.left}>
        <div className={styles.leftHeader}>
          <Text variant="h3">Zondex <Text variant="h3" as="span" color="accent">Шарага</Text></Text>
          <Text variant="caption" color="muted">Учись. Зарабатывай. Погашай долги.</Text>
        </div>

        <Text variant="overline" color="muted">Курсы</Text>

        <div className={styles.courses}>
          {COURSES.map(course => (
            <Card
              key={course.id}
              padding="md"
              hoverable
              highlighted={activeCourse?.id === course.id}
              onClick={() => handleCourseClick(course)}
              className={styles.courseCard}
            >
              <div className={styles.courseHeader}>
                <span className={styles.courseEmoji}>{course.emoji}</span>
                <div className={styles.courseMeta}>
                  <div className={styles.courseTitleRow}>
                    <Text variant="label">{course.title}</Text>
                    {course.free && <Chip variant="success" size="sm">БЕСПЛАТНО</Chip>}
                  </div>
                  <Text variant="caption" color="muted">{course.description}</Text>
                </div>
              </div>
              <div className={styles.courseFooter}>
                <Text variant="caption" color="muted" as="span">{course.lessons} уроков</Text>
                <Text variant="caption" color="accent" as="span">+{course.reward} лаве</Text>
              </div>
            </Card>
          ))}
        </div>
      </div>

      {/* Right — chat */}
      <div className={styles.right}>
        {/* Chat header */}
        <div className={styles.chatHeader}>
          <div className={styles.chatAiInfo}>
            <span className={styles.chatAiAvatar}>🤖</span>
            <div>
              <Text variant="label">MoshennikAI</Text>
              <Text variant="caption" color="success" as="span">онлайн</Text>
            </div>
          </div>
          <button className={styles.clearBtn} onClick={clearChat} title="Очистить чат">
            <Text variant="caption" color="muted">очистить</Text>
          </button>
        </div>

        {/* Messages */}
        <div className={styles.messages}>
          {messages.map(msg => (
            <div
              key={msg.id}
              className={[
                styles.message,
                msg.role === 'user' ? styles.messageUser : styles.messageAi,
              ].join(' ')}
            >
              {msg.role === 'ai' && (
                <span className={styles.messageAvatar}>🤖</span>
              )}
              <div className={styles.messageBubble}>
                <Text variant="body" as="span">{msg.text}</Text>
              </div>
            </div>
          ))}

          {/* Typing indicator */}
          {typing && (
            <div className={[styles.message, styles.messageAi].join(' ')}>
              <span className={styles.messageAvatar}>🤖</span>
              <div className={[styles.messageBubble, styles.typingBubble].join(' ')}>
                <span className={styles.typingDot} />
                <span className={styles.typingDot} />
                <span className={styles.typingDot} />
              </div>
            </div>
          )}

          <div ref={messagesEndRef} />
        </div>

        {/* Input */}
        <div className={styles.inputRow}>
          <textarea
            ref={inputRef}
            className={styles.textarea}
            placeholder="Задай вопрос MoshennikAI..."
            value={input}
            onChange={e => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            rows={1}
            disabled={typing}
          />
          <Button
            variant="primary"
            size="md"
            onClick={handleSend}
            disabled={!input.trim() || typing}
          >
            →
          </Button>
        </div>
        <Text variant="caption" color="muted" className={styles.inputHint}>
          Enter — отправить · Shift+Enter — новая строка
        </Text>
      </div>
    </div>
  )
}
