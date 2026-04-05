import { useNavigate } from 'react-router-dom'
import { Text, Card } from '../../ds/index.js'
import styles from './EduPage.module.css'

const EDU_SERVICES = [
  {
    id: 'chat',
    path: '/edu/chat',
    emoji: '🤖',
    title: 'MoshennikAI',
    description: 'AI-наставник по жизни на зоне. Задавай вопросы — получай ответы по понятиям.',
    badge: 'БЕСПЛАТНО',
  },
  {
    id: 'quiz',
    path: '/edu/quiz',
    emoji: '📝',
    title: 'Тест на понятия',
    description: 'Проверь знание тюремных понятий. За правильные ответы — лаве и респект.',
    badge: 'НАГРАДА',
  },
]

export default function EduPage() {
  const navigate = useNavigate()

  return (
    <div className={styles.launcherPage}>
      <div className={styles.launcherHeader}>
        <Text variant="h2">Zondex <Text variant="h2" as="span" color="accent">Шарага</Text></Text>
        <Text variant="body" color="muted">Учись. Зарабатывай. Погашай долги.</Text>
      </div>

      <div className={styles.launcherGrid}>
        {EDU_SERVICES.map(service => (
          <Card
            key={service.id}
            padding="lg"
            hoverable
            onClick={() => navigate(service.path)}
            className={styles.launcherCard}
          >
            <div className={styles.launcherEmoji}>{service.emoji}</div>
            <Text variant="h3">{service.title}</Text>
            <Text variant="body" color="muted">{service.description}</Text>
            <div className={styles.launcherArrow}>→</div>
          </Card>
        ))}
      </div>
    </div>
  )
}
