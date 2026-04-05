import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Text, Button, Card, Chip, ProgressBar, Currency } from '../../ds/index.js'
import useGamesStore from '../../store/useGamesStore.js'
import useLoanStore, { LOAN_AMOUNT, LOAN_BAN_LIMIT } from '../../store/useLoanStore.js'
import useUserStore from '../../store/useUserStore.js'
import { GAME_CATEGORIES } from '../../api/games.js'
import { RANKS, hasRankAccess } from '../../config/ranks.js'
import styles from './GamesPage.module.css'

// Таблица авторитетов — заглушка
const LEADERBOARD = [
  { id: 1, initials: 'КЗ', login: 'КолянЗэк',    balance:  12400 },
  { id: 2, initials: 'ВР', login: 'ВаняРулон',   balance:   8850 },
  { id: 3, initials: 'МТ', login: 'МаксТопор',   balance:   6100 },
  { id: 4, initials: 'СЧ', login: 'СаняЧёрная',  balance:  -2300 },
  { id: 5, initials: 'ПК', login: 'ПетяКасса',   balance:  -4100 },
]

// Форматирование таймера обратного отсчёта
function useCountdown(targetTs) {
  const [display, setDisplay] = useState('--:--')

  useEffect(() => {
    if (!targetTs) { setDisplay('--:--'); return }
    const tick = () => {
      const diff = Math.max(0, targetTs - Date.now())
      const m = String(Math.floor(diff / 60000)).padStart(2, '0')
      const s = String(Math.floor((diff % 60000) / 1000)).padStart(2, '0')
      setDisplay(`${m}:${s}`)
    }
    tick()
    const id = setInterval(tick, 1000)
    return () => clearInterval(id)
  }, [targetTs])

  return display
}

export default function GamesPage() {
  const navigate = useNavigate()

  // Games store
  const { games, activeCategory, filteredGames, fetchGames, setCategory, loading: gamesLoading } = useGamesStore()

  // Loan store
  const {
    debt, nextInterestAt, loading: loanLoading, borrow, debtPercent, isBanned,
  } = useLoanStore()

  // User store
  const { balance, respect, setBalance } = useUserStore()

  const countdown = useCountdown(nextInterestAt)

  useEffect(() => {
    fetchGames()
  }, [fetchGames])

  const handleBorrow = async () => {
    await borrow()
    // Зачисляем лаве на баланс пользователя
    setBalance(balance + LOAN_AMOUNT)
  }

  const filtered = activeCategory === 'Все игры'
    ? games
    : games.filter(g => g.category === activeCategory)

  return (
    <div className={styles.page}>
      {/* Left column */}
      <div className={styles.left}>
        {/* Games section */}
        <section className={styles.section}>
          <Text variant="overline" color="muted">Игровой зал</Text>

          {/* Category filters */}
          <div className={styles.filters}>
            {GAME_CATEGORIES.map(cat => (
              <Chip
                key={cat}
                variant="default"
                active={cat === activeCategory}
                style={{ cursor: 'pointer' }}
                onClick={() => setCategory(cat)}
              >
                {cat}
              </Chip>
            ))}
          </div>

          {/* Games grid */}
          {gamesLoading ? (
            <div className={styles.gamesGrid}>
              {[1, 2, 3].map(i => <div key={i} className={styles.gameSkeleton} />)}
            </div>
          ) : (
            <div className={styles.gamesGrid}>
              {filtered.map(game => {
                const accessible = hasRankAccess(respect, game.requiredRank)
                const requiredRank = game.requiredRank
                  ? RANKS.find(r => r.id === game.requiredRank)
                  : null
                return (
                  <GameCard
                    key={game.id}
                    game={game}
                    accessible={accessible}
                    requiredRank={requiredRank}
                    onClick={() => accessible && navigate(`/games/${game.id === 'shmon' ? 'shmon' : game.id}`)}
                  />
                )
              })}
            </div>
          )}
        </section>

        {/* Debt section */}
        {debt < 0 && (
          <section className={styles.section}>
            <Text variant="overline" color="muted">Состояние долга</Text>
            <Card padding="md" className={styles.debtCard}>
              <div className={styles.debtHeader}>
                <span className={styles.debtDot} />
                <Text variant="label" color="danger">Долг смотрящему</Text>
              </div>
              <div className={styles.debtAmount}>
                <Text variant="h2" color="danger" as="span">{debt.toLocaleString('ru')}</Text>
                <Text variant="h2" color="danger" as="span"> лаве</Text>
              </div>
              <Text variant="caption" color="muted">
                Следующие проценты через: <Text variant="caption" color="accent" as="span">{countdown}</Text>
              </Text>
              <ProgressBar value={debtPercent} theme="danger" size="sm" />
              <Text variant="caption" color="muted">
                Лимит бана: <Text variant="caption" color="danger" as="span">{LOAN_BAN_LIMIT.toLocaleString('ru')} лаве</Text>.{' '}
                Сейчас {debtPercent}% от лимита. Каждые 10 мин +5% от суммы долга.
              </Text>
            </Card>
          </section>
        )}

        {/* Ban screen preview */}
        {isBanned && (
          <Card padding="lg" className={styles.banCard}>
            <div className={styles.banInner}>
              <span className={styles.banLock}>🔒</span>
              <Text variant="label" color="danger">ЭТАПИРОВАН</Text>
              <Text variant="caption" color="muted">Так выглядит бан-экран при {LOAN_BAN_LIMIT.toLocaleString('ru')}</Text>
              <Text variant="h3" color="danger">{countdown}</Text>
            </div>
          </Card>
        )}
      </div>

      {/* Right column */}
      <div className={styles.right}>
        {/* Loan block */}
        <Card padding="lg">
          <Text variant="label">Взять кредит</Text>
          <Button
            variant="secondary"
            fullWidth
            onClick={handleBorrow}
            disabled={loanLoading || isBanned}
          >
            Занять {LOAN_AMOUNT} лаве
          </Button>
          <div className={styles.loanMeta}>
            <Text variant="caption" color="muted">+5% каждые 10 минут</Text>
            <Text variant="caption" color="muted">Лимит бана: {LOAN_BAN_LIMIT.toLocaleString('ru')} лаве</Text>
          </div>
          <div className={styles.loanSpin}>
            <Text variant="caption" color="muted">Ежедневный спин</Text>
            <Text variant="caption" color="accent">доступен</Text>
          </div>
        </Card>

        {/* Leaderboard */}
        <Card padding="md">
          <Text variant="label">Таблица авторитетов</Text>
          <div className={styles.leaderboard}>
            {LEADERBOARD.map((entry, idx) => (
              <div key={entry.id} className={styles.leaderRow}>
                <Text variant="caption" color="muted" as="span" className={styles.leaderRank}>
                  {idx + 1}
                </Text>
                <span className={styles.leaderAvatar}>{entry.initials}</span>
                <Text variant="caption" as="span" className={styles.leaderLogin}>
                  {entry.login}
                </Text>
                <Currency
                  amount={Math.abs(entry.balance).toLocaleString('ru')}
                  unit=""
                  size="sm"
                  color={entry.balance >= 0 ? 'default' : 'danger'}
                  className={styles.leaderBalance}
                />
              </div>
            ))}
          </div>
        </Card>

        {/* Ban preview card */}
        <Card padding="lg" className={styles.banPreviewCard}>
          <div className={styles.banInner}>
            <span className={styles.banLock}>🔒</span>
            <Text variant="label" color="danger">ЭТАПИРОВАН</Text>
            <Text variant="caption" color="muted">Так выглядит бан-экран при {LOAN_BAN_LIMIT.toLocaleString('ru')}</Text>
            <Text variant="h3" color="danger">11:47:32</Text>
          </div>
        </Card>
      </div>
    </div>
  )
}

function GameCard({ game, accessible = true, requiredRank, onClick }) {
  const badgeVariant = {
    'ХИТ':       'warn',
    'НОВОЕ':      'info',
    'БЕСПЛАТНО':  'success',
  }[game.badge] ?? 'default'

  return (
    <Card
      padding="md"
      hoverable={accessible}
      onClick={accessible ? onClick : undefined}
      className={[styles.gameCard, !accessible ? styles.gameCardLocked : ''].filter(Boolean).join(' ')}
    >
      {/* Lock overlay для недоступных игр */}
      {!accessible && (
        <div className={styles.gameLockOverlay}>
          <span>🔒</span>
          {requiredRank && (
            <Text variant="caption" color="muted" as="span">
              <Text variant="caption" as="span" style={{ color: requiredRank.color }}>
                {requiredRank.emoji} {requiredRank.title}
              </Text>
            </Text>
          )}
        </div>
      )}
      {game.badge && accessible && (
        <div className={styles.gameBadge}>
          <Chip variant={badgeVariant} size="sm">{game.badge}</Chip>
        </div>
      )}
      <div className={styles.gameEmoji}>{game.emoji}</div>
      <Text variant="label">{game.title}</Text>
      <Text variant="caption" color="muted">{game.description}</Text>
    </Card>
  )
}
