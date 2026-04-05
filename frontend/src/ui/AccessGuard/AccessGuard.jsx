import { useNavigate } from 'react-router-dom'
import { Text, Button, Card, ProgressBar } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import { hasAccess, getRequiredRank } from '../../config/access.js'
import { getRankByRespect, getRankProgress } from '../../config/ranks.js'
import styles from './AccessGuard.module.css'

/**
 * AccessGuard — обёртка для защиты страниц по рангу
 *
 * @param {string} sectionId — ключ из SECTION_ACCESS
 * @param {React.ReactNode} children
 */
export default function AccessGuard({ sectionId, children }) {
  const { respect } = useUserStore()
  const navigate = useNavigate()

  if (hasAccess(sectionId, respect)) {
    return children
  }

  const requiredRank = getRequiredRank(sectionId)
  const currentRank = getRankByRespect(respect)
  const progress = getRankProgress(respect)
  const needed = requiredRank ? requiredRank.minRespect - respect : 0

  return (
    <div className={styles.page}>
      <Card padding="lg" className={styles.card}>
        <div className={styles.lockIcon}>🔒</div>

        <Text variant="h2" color="danger">Доступ закрыт</Text>

        <Text variant="body" color="muted">
          Этот раздел доступен только с ранга{' '}
          <Text variant="body" as="span" style={{ color: requiredRank?.color }}>
            {requiredRank?.emoji} {requiredRank?.title}
          </Text>
        </Text>

        <div className={styles.rankInfo}>
          <div className={styles.rankRow}>
            <div className={styles.rankItem}>
              <Text variant="overline" color="muted">Твой ранг</Text>
              <Text variant="h3" as="span" style={{ color: currentRank.color }}>
                {currentRank.emoji} {currentRank.title}
              </Text>
              <Text variant="caption" color="muted">{respect} респекта</Text>
            </div>
            <div className={styles.rankArrow}>→</div>
            <div className={styles.rankItem}>
              <Text variant="overline" color="muted">Нужен ранг</Text>
              <Text variant="h3" as="span" style={{ color: requiredRank?.color, opacity: 0.6 }}>
                {requiredRank?.emoji} {requiredRank?.title}
              </Text>
              <Text variant="caption" color="muted">от {requiredRank?.minRespect} респекта</Text>
            </div>
          </div>

          <div className={styles.progressSection}>
            <div className={styles.progressLabels}>
              <Text variant="caption" color="muted" as="span">Прогресс до {requiredRank?.title}</Text>
              <Text variant="caption" color="accent" as="span">ещё {needed} респекта</Text>
            </div>
            <ProgressBar value={progress} theme="warn" size="md" />
          </div>
        </div>

        <Text variant="caption" color="muted" className={styles.hint}>
          Зарабатывай респект в Шараге, играх и барахолке
        </Text>

        <div className={styles.actions}>
          <Button variant="primary" onClick={() => navigate('/edu')}>
            Идти в Шарагу
          </Button>
          <Button variant="secondary" onClick={() => navigate('/ranks')}>
            Все звания
          </Button>
          <Button variant="ghost" onClick={() => navigate('/')}>
            На главную
          </Button>
        </div>
      </Card>
    </div>
  )
}
