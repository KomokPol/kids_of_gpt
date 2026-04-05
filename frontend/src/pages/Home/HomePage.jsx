import { useNavigate } from 'react-router-dom'
import { Text, Card, Chip } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import { getRankByRespect } from '../../config/ranks.js'
import { hasAccess, getRequiredRank } from '../../config/access.js'
import styles from './HomePage.module.css'

const SERVICES = [
  {
    id: 'edu',
    path: '/edu',
    emoji: '🤖',
    title: 'Шарага',
    description: 'AI-наставник по жизни на зоне. Курсы, понятия, законы.',
    badge: 'БЕСПЛАТНО',
    badgeVariant: 'success',
  },
  {
    id: 'games',
    path: '/games',
    emoji: '🎰',
    title: 'Игровой зал',
    description: 'Слоты, карты, рулетка. Шмон, Смотрящий, Приговор.',
    badge: 'ГОРЯЧО',
    badgeVariant: 'warn',
  },
  {
    id: 'market',
    path: '/market',
    emoji: '🛒',
    title: 'Барахолка',
    description: 'Маркетплейс зоны. Еда, шмотки, инструменты, статус.',
    badge: null,
    badgeVariant: null,
  },
  {
    id: 'ranks',
    path: '/ranks',
    emoji: '🏅',
    title: 'Звания',
    description: 'Все звания зоны. Привилегии, требования, прогресс.',
    badge: null,
    badgeVariant: null,
  },
  {
    id: 'catalog',
    path: '/catalog',
    emoji: '📼',
    title: 'Видеомагнитофон',
    description: 'Только хтонь. Только правда. Каталог фильмов.',
    badge: null,
    badgeVariant: null,
  },
]

export default function HomePage() {
  const navigate = useNavigate()
  const { respect, balance, login } = useUserStore()
  const rank = getRankByRespect(respect)

  return (
    <div className={styles.page}>
      {/* Hero */}
      <section className={styles.hero}>
        <div className={styles.heroText}>
          <Text variant="h1">Zondax</Text>
          <Text variant="body" color="muted">
            Платформа зоны. Учись, играй, торгуй, смотри.
          </Text>
        </div>
        {login && (
          <div className={styles.heroUser}>
            <span className={styles.heroRankEmoji}>{rank.emoji}</span>
            <div>
              <Text variant="label" style={{ color: rank.color }}>{rank.title}</Text>
              <div className={styles.heroStats}>
                <Text variant="caption" color="muted" as="span">
                  <Text variant="caption" style={{ color: 'var(--color-respect)' }} as="span">{respect}</Text> респект
                </Text>
                <Text variant="caption" color="muted" as="span">·</Text>
                <Text variant="caption" color="muted" as="span">
                  <Text variant="caption" color="accent" as="span">{balance}</Text> лаве
                </Text>
              </div>
            </div>
          </div>
        )}
      </section>

      {/* Services grid */}
      <section className={styles.services}>
        <Text variant="overline" color="muted">Сервисы</Text>
        <div className={styles.grid}>
          {SERVICES.map(service => {
            const accessible = hasAccess(service.id, respect)
            const requiredRank = getRequiredRank(service.id)
            return (
              <ServiceCard
                key={service.id}
                service={service}
                accessible={accessible}
                requiredRank={requiredRank}
                onClick={() => navigate(service.path)}
              />
            )
          })}
        </div>
      </section>
    </div>
  )
}

function ServiceCard({ service, accessible, requiredRank, onClick }) {
  return (
    <Card
      padding="lg"
      hoverable={accessible}
      onClick={accessible ? onClick : undefined}
      className={[
        styles.serviceCard,
        !accessible ? styles.serviceCardLocked : '',
      ].filter(Boolean).join(' ')}
    >
      {/* Lock overlay */}
      {!accessible && (
        <div className={styles.lockOverlay}>
          <span className={styles.lockIcon}>🔒</span>
          <Text variant="caption" color="muted" as="span">
            Нужен ранг:{' '}
            <Text variant="caption" as="span" style={{ color: requiredRank?.color }}>
              {requiredRank?.emoji} {requiredRank?.title}
            </Text>
          </Text>
        </div>
      )}

      {service.badge && accessible && (
        <div className={styles.cardBadge}>
          <Chip variant={service.badgeVariant} size="sm">{service.badge}</Chip>
        </div>
      )}

      <div className={styles.cardEmoji} style={{ opacity: accessible ? 1 : 0.3 }}>
        {service.emoji}
      </div>
      <Text variant="h3" className={styles.cardTitle} color={accessible ? 'default' : 'dim'}>
        {service.title}
      </Text>
      <Text variant="caption" color={accessible ? 'muted' : 'dim'}>
        {service.description}
      </Text>

      {accessible && <div className={styles.cardArrow}>→</div>}
    </Card>
  )
}
