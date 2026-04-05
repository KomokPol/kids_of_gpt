import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Text, Button, Card, ProgressBar } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import { getQuizQuestions, calcQuizReward } from '../../api/quiz.js'
import styles from './EduQuizPage.module.css'

const TOTAL_QUESTIONS = 5

export default function EduQuizPage() {
  const navigate = useNavigate()
  const { balance, respect, setBalance, setRespect } = useUserStore()

  const [phase, setPhase] = useState('loading') // loading | quiz | result
  const [questions, setQuestions] = useState([])
  const [currentIdx, setCurrentIdx] = useState(0)
  const [selected, setSelected] = useState(null)       // индекс выбранного ответа
  const [answered, setAnswered] = useState(false)      // показать правильный ответ
  const [correctCount, setCorrectCount] = useState(0)
  const [answers, setAnswers] = useState([])           // история ответов

  useEffect(() => {
    getQuizQuestions(TOTAL_QUESTIONS).then(qs => {
      setQuestions(qs)
      setPhase('quiz')
    })
  }, [])

  const currentQ = questions[currentIdx]
  const progress = questions.length > 0 ? Math.round((currentIdx / questions.length) * 100) : 0

  const handleSelect = (idx) => {
    if (answered) return
    setSelected(idx)
    setAnswered(true)
    const isCorrect = idx === currentQ.correct
    if (isCorrect) setCorrectCount(c => c + 1)
    setAnswers(prev => [...prev, { questionId: currentQ.id, selected: idx, correct: currentQ.correct, isCorrect }])
  }

  const handleNext = () => {
    if (currentIdx + 1 >= questions.length) {
      // Финиш — начислить награду
      const reward = calcQuizReward(correctCount + (selected === currentQ.correct ? 1 : 0), questions.length)
      setBalance(balance + reward.lave)
      setRespect(respect + reward.respect)
      setPhase('result')
    } else {
      setCurrentIdx(i => i + 1)
      setSelected(null)
      setAnswered(false)
    }
  }

  const handleRestart = () => {
    setPhase('loading')
    setCurrentIdx(0)
    setSelected(null)
    setAnswered(false)
    setCorrectCount(0)
    setAnswers([])
    getQuizQuestions(TOTAL_QUESTIONS).then(qs => {
      setQuestions(qs)
      setPhase('quiz')
    })
  }

  if (phase === 'loading') {
    return (
      <div className={styles.page}>
        <div className={styles.loading}>
          <Text variant="h3" color="muted">Загружаем вопросы...</Text>
        </div>
      </div>
    )
  }

  if (phase === 'result') {
    const total = questions.length
    const reward = calcQuizReward(correctCount, total)
    const ratio = correctCount / total

    return (
      <div className={styles.page}>
        <Card padding="lg" className={styles.resultCard}>
          <div className={styles.resultEmoji}>
            {ratio >= 0.8 ? '🏆' : ratio >= 0.5 ? '👍' : '📚'}
          </div>
          <Text variant="h2">
            {ratio >= 0.8 ? 'Отлично!' : ratio >= 0.5 ? 'Неплохо' : 'Учись дальше'}
          </Text>
          <Text variant="body" color="muted">
            Правильных ответов: <Text variant="body" color="accent" as="span">{correctCount}</Text> из {total}
          </Text>

          <div className={styles.rewardBlock}>
            <Text variant="overline" color="muted">Награда</Text>
            <div className={styles.rewardRow}>
              <Text variant="h3" color="accent">+{reward.lave} лаве</Text>
              <Text variant="h3" style={{ color: 'var(--color-respect)' }}>+{reward.respect} респект</Text>
            </div>
          </div>

          {/* Разбор ответов */}
          <div className={styles.answerReview}>
            <Text variant="overline" color="muted">Разбор</Text>
            {questions.map((q, i) => {
              const ans = answers[i]
              return (
                <div key={q.id} className={[styles.reviewItem, ans?.isCorrect ? styles.reviewCorrect : styles.reviewWrong].join(' ')}>
                  <Text variant="caption" as="span">{ans?.isCorrect ? '✓' : '✗'} {q.question}</Text>
                </div>
              )
            })}
          </div>

          <div className={styles.resultActions}>
            <Button variant="primary" onClick={handleRestart}>Пройти ещё раз</Button>
            <Button variant="secondary" onClick={() => navigate('/edu')}>В Шарагу</Button>
          </div>
        </Card>
      </div>
    )
  }

  return (
    <div className={styles.page}>
      <div className={styles.quizHeader}>
        <button className={styles.backBtn} onClick={() => navigate('/edu')}>← Шарага</button>
        <Text variant="caption" color="muted">Вопрос {currentIdx + 1} из {questions.length}</Text>
      </div>

      <ProgressBar value={progress} theme="default" size="sm" />

      <Card padding="lg" className={styles.questionCard}>
        <Text variant="h3">{currentQ.question}</Text>

        <div className={styles.options}>
          {currentQ.options.map((opt, idx) => {
            let optClass = styles.option
            if (answered) {
              if (idx === currentQ.correct) optClass = [styles.option, styles.optionCorrect].join(' ')
              else if (idx === selected && idx !== currentQ.correct) optClass = [styles.option, styles.optionWrong].join(' ')
            } else if (idx === selected) {
              optClass = [styles.option, styles.optionSelected].join(' ')
            }
            return (
              <button key={idx} className={optClass} onClick={() => handleSelect(idx)} disabled={answered}>
                <span className={styles.optionLetter}>{String.fromCharCode(65 + idx)}</span>
                <Text variant="body" as="span">{opt}</Text>
              </button>
            )
          })}
        </div>

        {answered && (
          <div className={styles.explanation}>
            <Text variant="caption" color={selected === currentQ.correct ? 'success' : 'danger'} as="span">
              {selected === currentQ.correct ? '✓ Правильно!' : '✗ Неверно.'}
            </Text>
            <Text variant="caption" color="muted"> {currentQ.explanation}</Text>
          </div>
        )}

        {answered && (
          <Button variant="primary" fullWidth onClick={handleNext}>
            {currentIdx + 1 >= questions.length ? 'Завершить' : 'Следующий вопрос →'}
          </Button>
        )}
      </Card>
    </div>
  )
}
