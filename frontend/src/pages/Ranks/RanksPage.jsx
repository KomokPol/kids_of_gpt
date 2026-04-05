import { Text, Card, ProgressBar, Chip } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import { RANKS, getRankByRespect, getRankProgress, getNextRank } from '../../config/ranks.js'
import styles from './RanksPage.module.css'

// Привилегии каждого звания
const RANK_PERKS = {
  'новичок':    ['Доступ к базовым товарам Барахолки', 'Бесплатные курсы MoshennikAI', 'Игра Шмон'],
  'мужик':      ['Кирзачи и игла в Барахолке', 'Расширенный доступ к курсам', 'Повышенные ставки в слотах'],
  'бродяга':    ['Телогрейка и татуировка', 'Курс "Выживание на зоне"', 'Доступ к редким товарам'],
  'авторитет':  ['Радиоприёмник и перстень', 'Все курсы MoshennikAI', 'Особый статус в таблице авторитетов'],
  'положенец':  ['Эксклюзивные товары', 'Сниженные проценты по долгу', 'Приоритет в очереди'],
  'смотрящий':  ['Корона в Барахолке', 'Управление общаком', 'Особый значок в профиле'],
  'вор':        ['Все привилегии', 'Неприкосновенность', 'Легенда зоны'],
}

export default function RanksPage() {
  const { respect } = useUserStore()
  const currentRank = getRankByRespect(respect)
  const nextRank = getNextRank(respect)
  const progress = getRankProgress(respect)

  return (
    <div className={styles.page}>
      {/* Header */}
      <div className={styles.pageHeader}>
        <Text variant="h2">Звания</Text>
        <Text variant="body" color="muted">
          Зарабатывай респект — повышай звание. Чем выше звание, тем больше привилегий.
        </Text>
      </div>

      {/* Current rank hero */}
      <Card padding="lg" className={styles.currentCard}>
        <div className={styles.currentInner}>
          <div className={styles.currentLeft}>
            <Text variant="overline" color="muted">Твоё звание</Text>
            <div className={styles.currentRank}>
              <span className={styles.currentEmoji}>{currentRank.emoji}</span>
              <Text variant="h2" as="span" style={{ color: currentRank.color }}>
                {currentRank.title}
              </Text>
            </div>
            <Text variant="body" color="muted">
              {respect} респекта
            </Text>
          </div>

          {nextRank ? (
            <div className={styles.currentRight}>
              <div className={styles.progressLabels}>
                <Text variant="caption" color="muted" as="span">
                  {currentRank.emoji} {currentRank.title}
                </Text>
                <Text variant="caption" color="muted" as="span">
                  {nextRank.emoji} {nextRank.title}
                </Text>
              </div>
              <ProgressBar value={progress} theme="default" size="lg" />
              <div className={styles.progressInfo}>
                <Text variant="caption" color="muted" as="span">{progress}% до следующего</Text>
                <Text variant="caption" color="accent" as="span">
                  ещё {nextRank.minRespect - respect} респекта
                </Text>
              </div>
            </div>
          ) : (
            <div className={styles.currentRight}>
              <Text variant="h3" color="accent">Максимальное звание 👑</Text>
              <Text variant="caption" color="muted">Ты достиг вершины. Вор в законе.</Text>
            </div>
          )}
        </div>
      </Card>

      {/* All ranks */}
      <div className={styles.ranksList}>
        {RANKS.map((rank, idx) => {
          const unlocked = respect >= rank.minRespect
          const isCurrent = rank.id === currentRank.id
          const perks = RANK_PERKS[rank.id] ?? []

          return (
            <Card
              key={rank.id}
              padding="lg"
              highlighted={isCurrent}
              className={[
                styles.rankCard,
                !unlocked ? styles.rankCardLocked : '',
                isCurrent ? styles.rankCardCurrent : '',
              ].filter(Boolean).join(' ')}
            >
              <div className={styles.rankRow}>
                {/* Left: emoji + info */}
                <div className={styles.rankLeft}>
                  <div className={styles.rankNum}>
                    <Text variant="caption" color="muted" as="span">{idx + 1}</Text>
                  </div>
                  <span
                    className={styles.rankEmoji}
                    style={{ opacity: unlocked ? 1 : 0.35 }}
                  >
                    {rank.emoji}
                  </span>
                  <div className={styles.rankInfo}>
                    <div className={styles.rankTitleRow}>
                      <Text
                        variant="h3"
                        as="span"
                        style={{ color: unlocked ? rank.color : 'var(--color-text-dim)' }}
                      >
                        {rank.title}
                      </Text>
                      {isCurrent && (
                        <Chip variant="success" size="sm">Текущее</Chip>
                      )}
                      {!unlocked && (
                        <Chip variant="default" size="sm">🔒 Закрыто</Chip>
                      )}
                    </div>
                    <Text variant="caption" color="muted">
                      от {rank.minRespect.toLocaleString('ru')} респекта
                    </Text>
                  </div>
                </div>

                {/* Right: perks */}
                <div className={styles.rankPerks}>
                  {perks.map((perk, i) => (
                    <div key={i} className={styles.perkItem}>
                      <span className={styles.perkDot} style={{
                        backgroundColor: unlocked ? rank.color : 'var(--color-text-dim)'
                      }} />
                      <Text
                        variant="caption"
                        color={unlocked ? 'default' : 'dim'}
                        as="span"
                      >
                        {perk}
                      </Text>
                    </div>
                  ))}
                </div>
              </div>

              {/* Progress bar for current rank */}
              {isCurrent && nextRank && (
                <div className={styles.rankProgress}>
                  <ProgressBar value={progress} theme="default" size="sm" />
                </div>
              )}
            </Card>
          )
        })}
      </div>
    </div>
  )
}
