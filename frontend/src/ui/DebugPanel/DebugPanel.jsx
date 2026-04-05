import { useState } from 'react'
import { Text, Button } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import { RANKS, getRankByRespect } from '../../config/ranks.js'
import styles from './DebugPanel.module.css'

/**
 * DebugPanel — панель отладки (только для разработки)
 * Открывается по кнопке 🐛 снизу слева
 */
export default function DebugPanel() {
  const [isOpen, setIsOpen] = useState(false)
  const { respect, balance, login, setRespect, setBalance } = useUserStore()
  const currentRank = getRankByRespect(respect)

  const [respectInput, setRespectInput] = useState('')
  const [balanceInput, setBalanceInput] = useState('')

  const applyRespect = () => {
    const val = parseInt(respectInput, 10)
    if (!isNaN(val) && val >= 0) {
      setRespect(val)
      setRespectInput('')
    }
  }

  const applyBalance = () => {
    const val = parseInt(balanceInput, 10)
    if (!isNaN(val)) {
      setBalance(val)
      setBalanceInput('')
    }
  }

  const setRank = (rank) => {
    setRespect(rank.minRespect)
  }

  if (!isOpen) {
    return (
      <button className={styles.toggleBtn} onClick={() => setIsOpen(true)} title="Debug Panel">
        🐛
      </button>
    )
  }

  return (
    <div className={styles.panel}>
      <div className={styles.header}>
        <Text variant="label" color="accent">🐛 Debug Panel</Text>
        <button className={styles.closeBtn} onClick={() => setIsOpen(false)}>✕</button>
      </div>

      {/* Current state */}
      <div className={styles.section}>
        <Text variant="overline" color="muted">Текущее состояние</Text>
        <div className={styles.stateRow}>
          <Text variant="caption" color="muted" as="span">Логин:</Text>
          <Text variant="caption" as="span">{login ?? '—'}</Text>
        </div>
        <div className={styles.stateRow}>
          <Text variant="caption" color="muted" as="span">Ранг:</Text>
          <Text variant="caption" as="span" style={{ color: currentRank.color }}>
            {currentRank.emoji} {currentRank.title}
          </Text>
        </div>
        <div className={styles.stateRow}>
          <Text variant="caption" color="muted" as="span">Респект:</Text>
          <Text variant="caption" style={{ color: 'var(--color-respect)' }} as="span">{respect}</Text>
        </div>
        <div className={styles.stateRow}>
          <Text variant="caption" color="muted" as="span">Баланс:</Text>
          <Text variant="caption" color="accent" as="span">{balance} лаве</Text>
        </div>
      </div>

      {/* Set respect */}
      <div className={styles.section}>
        <Text variant="overline" color="muted">Установить респект</Text>
        <div className={styles.inputRow}>
          <input
            className={styles.input}
            type="number"
            min="0"
            placeholder="Значение..."
            value={respectInput}
            onChange={e => setRespectInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && applyRespect()}
          />
          <Button variant="primary" size="sm" onClick={applyRespect}>OK</Button>
        </div>
      </div>

      {/* Set balance */}
      <div className={styles.section}>
        <Text variant="overline" color="muted">Установить баланс</Text>
        <div className={styles.inputRow}>
          <input
            className={styles.input}
            type="number"
            placeholder="Лаве..."
            value={balanceInput}
            onChange={e => setBalanceInput(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && applyBalance()}
          />
          <Button variant="primary" size="sm" onClick={applyBalance}>OK</Button>
        </div>
      </div>

      {/* Quick rank presets */}
      <div className={styles.section}>
        <Text variant="overline" color="muted">Быстрый ранг</Text>
        <div className={styles.rankBtns}>
          {RANKS.map(rank => (
            <button
              key={rank.id}
              className={[
                styles.rankBtn,
                currentRank.id === rank.id ? styles.rankBtnActive : '',
              ].filter(Boolean).join(' ')}
              style={{ '--rank-color': rank.color }}
              onClick={() => setRank(rank)}
              title={`${rank.minRespect} респекта`}
            >
              {rank.emoji} {rank.title}
            </button>
          ))}
        </div>
      </div>

      {/* Quick balance presets */}
      <div className={styles.section}>
        <Text variant="overline" color="muted">Быстрый баланс</Text>
        <div className={styles.presetBtns}>
          {[0, 100, 500, 1000, 5000].map(val => (
            <button
              key={val}
              className={styles.presetBtn}
              onClick={() => setBalance(val)}
            >
              {val}
            </button>
          ))}
        </div>
      </div>
    </div>
  )
}
