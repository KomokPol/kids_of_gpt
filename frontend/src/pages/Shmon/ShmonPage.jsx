import { useEffect } from 'react'
import { Text, Button, Card, Chip, Currency } from '../../ds/index.js'
import useShmonStore, { SYMBOLS, PAYOUTS, MIN_BET, MAX_BET } from '../../store/useShmonStore.js'
import useUserStore from '../../store/useUserStore.js'
import useLoanStore, { LOAN_AMOUNT, LOAN_BAN_LIMIT } from '../../store/useLoanStore.js'
import styles from './ShmonPage.module.css'

// Таблица авторитетов — заглушка
const LEADERBOARD = [
  { id: 1, initials: 'КЗ', login: 'КолянЗэк',   balance:  12400 },
  { id: 2, initials: 'ВР', login: 'ВаняРулон',  balance:   8850 },
  { id: 3, initials: 'МТ', login: 'МаксТопор',  balance:   6100 },
  { id: 4, initials: 'СЧ', login: 'СаняЧёрная', balance:  -2300 },
  { id: 5, initials: 'ПК', login: 'ПетяКасса',  balance:  -4100 },
]

const SYMBOL_MAP = Object.fromEntries(SYMBOLS.map(s => [s.id, s]))

export default function ShmonPage() {
  const {
    reels, spinning, bet, lastWin, lastResult,
    wins, losses, record, respectGained,
    spin, increaseBet, decreaseBet,
  } = useShmonStore()

  const { balance, setBalance, respect, setRespect } = useUserStore()
  const { debt, borrow, debtPercent } = useLoanStore()

  const handleSpin = async () => {
    if (spinning || balance < bet) return
    const result = await spin((delta) => {
      setBalance(balance + delta)
    })
    if (result?.respectDelta > 0) {
      setRespect(respect + result.respectDelta)
    }
  }

  const handleBorrow = async () => {
    await borrow()
    setBalance(balance + LOAN_AMOUNT)
  }

  return (
    <div className={styles.page}>
      {/* Left — game */}
      <div className={styles.left}>
        <Card padding="lg" className={styles.gameCard}>
          {/* Title */}
          <div className={styles.gameTitle}>
            <Text variant="h3">Шмон</Text>
            <Chip variant="warn" size="sm">СЛОТЫ</Chip>
          </div>

          {/* Reels */}
          <div className={styles.reels}>
            {reels.map((symbolId, i) => {
              const sym = SYMBOL_MAP[symbolId]
              return (
                <div
                  key={i}
                  className={[
                    styles.reel,
                    spinning ? styles.reelSpinning : '',
                    lastResult === 'win' && !spinning ? styles.reelWin : '',
                  ].filter(Boolean).join(' ')}
                >
                  <span className={styles.reelEmoji}>{sym?.emoji ?? '🔒'}</span>
                </div>
              )
            })}
          </div>

          {/* Win message */}
          {lastWin !== null && !spinning && (
            <div className={[styles.winMsg, lastResult === 'win' ? styles.winMsgWin : styles.winMsgLose].join(' ')}>
              {lastResult === 'win'
                ? <Text variant="label" color="success">+{lastWin} лаве!</Text>
                : <Text variant="caption" color="muted">Ставь лаве — крути барабан</Text>
              }
            </div>
          )}
          {lastWin === null && (
            <Text variant="caption" color="muted" className={styles.hint}>
              Ставь лаве — крути барабан
            </Text>
          )}

          {/* Bet controls */}
          <div className={styles.betRow}>
            <Text variant="label" color="muted" as="span">СТАВКА</Text>
            <div className={styles.betControls}>
              <button
                className={styles.betBtn}
                onClick={decreaseBet}
                disabled={bet <= MIN_BET || spinning}
              >−</button>
              <Text variant="label" as="span" className={styles.betValue}>{bet}</Text>
              <button
                className={styles.betBtn}
                onClick={increaseBet}
                disabled={bet >= MAX_BET || spinning}
              >+</button>
            </div>
            <Text variant="caption" color="muted" as="span">лаве</Text>
          </div>

          {/* Spin button */}
          <Button
            variant="secondary"
            size="lg"
            fullWidth
            onClick={handleSpin}
            disabled={spinning || balance < bet}
          >
            {spinning ? 'КРУТИТСЯ...' : 'КРУТИТЬ'}
          </Button>

          {/* Payouts table */}
          <div className={styles.payouts}>
            <Text variant="overline" color="muted">Выплаты</Text>
            {PAYOUTS.map((p, i) => (
              <div key={i} className={styles.payoutRow}>
                <span className={styles.payoutLabel}>{p.label}</span>
                <Text variant="caption" color="muted" as="span">{p.desc}</Text>
              </div>
            ))}
          </div>
        </Card>

        {/* Session stats */}
        <div className={styles.stats}>
          {[
            { label: 'ВЫИГРЫШИ',        value: wins },
            { label: 'ПРОИГРЫШИ',       value: losses },
            { label: 'РЕКОРД',          value: record },
            { label: 'РЕСПЕКТ ЗА СЕССИЮ', value: `+${respectGained}`, accent: true },
          ].map(stat => (
            <div key={stat.label} className={styles.statItem}>
              <Text variant="overline" color="muted">{stat.label}</Text>
              <Text
                variant="h3"
                as="span"
                style={{ color: stat.accent ? 'var(--color-respect)' : undefined }}
              >
                {stat.value}
              </Text>
            </div>
          ))}
        </div>
      </div>

      {/* Right — sidebar */}
      <div className={styles.right}>
        {/* Daily spin */}
        <Card padding="md" className={styles.dailyCard}>
          <Text variant="caption" color="accent">Ежедневный бесплатный спин</Text>
          <Button variant="primary" fullWidth size="md">Получить спин</Button>
        </Card>

        {/* Debt block */}
        <Card padding="md">
          <Text variant="label">Долг смотрящему</Text>
          <div className={styles.debtRow}>
            <Text
              variant="label"
              as="span"
              style={{ color: debt < 0 ? 'var(--color-danger)' : 'var(--color-text-muted)' }}
            >
              {debt === 0 ? '0' : debt.toLocaleString('ru')} лаве
            </Text>
            <Text variant="caption" color="muted" as="span">
              лимит: {LOAN_BAN_LIMIT.toLocaleString('ru')}
            </Text>
          </div>
          {debt < 0
            ? <Text variant="caption" color="muted">Долга нет</Text>
            : null
          }
          <Button
            variant="secondary"
            fullWidth
            onClick={handleBorrow}
          >
            Занять {LOAN_AMOUNT} лаве
          </Button>
          <Text variant="caption" color="muted">
            +5% каждые 10 сек · бан при {LOAN_BAN_LIMIT.toLocaleString('ru')}
          </Text>
        </Card>

        {/* Leaderboard */}
        <Card padding="md">
          <Text variant="label">Авторитеты недели</Text>
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
                <Text
                  variant="caption"
                  as="span"
                  style={{ color: entry.balance >= 0 ? 'var(--color-text)' : 'var(--color-danger)', flexShrink: 0 }}
                >
                  {entry.balance >= 0 ? '' : ''}{Math.abs(entry.balance).toLocaleString('ru')}
                </Text>
              </div>
            ))}
          </div>
        </Card>
      </div>
    </div>
  )
}
