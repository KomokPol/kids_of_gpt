import { useEffect, useState } from 'react'
import { Text, Button, Card, Chip, ProgressBar } from '../../ds/index.js'
import useMarketStore from '../../store/useMarketStore.js'
import useUserStore from '../../store/useUserStore.js'
import { MARKET_CATEGORIES } from '../../api/market.js'
import { RANKS, getRankByRespect, getNextRank, getRankProgress, hasRankAccess } from '../../config/ranks.js'
import styles from './MarketPage.module.css'

export default function MarketPage() {
  const { items, activeCategory, loading, fetchItems, setCategory, buyItem, purchased } = useMarketStore()
  const { balance, respect, setBalance } = useUserStore()
  const [notification, setNotification] = useState(null)

  useEffect(() => { fetchItems() }, [fetchItems])

  const currentRank = getRankByRespect(respect)
  const nextRank = getNextRank(respect)
  const rankProgress = getRankProgress(respect)

  const filtered = activeCategory === 'Всё'
    ? items
    : items.filter(i => i.category === activeCategory)

  const showNotification = (msg, type = 'success') => {
    setNotification({ msg, type })
    setTimeout(() => setNotification(null), 2500)
  }

  const handleBuy = (item) => {
    if (balance < item.price) {
      showNotification('Недостаточно лаве', 'error')
      return
    }
    if (!hasRankAccess(respect, item.requiredRank)) {
      showNotification(`Нужно звание: ${RANKS.find(r => r.id === item.requiredRank)?.title}`, 'error')
      return
    }
    const result = buyItem(item.id, (delta) => setBalance(balance + delta))
    if (result === 'ok') {
      showNotification(`${item.emoji} ${item.title} куплено!`)
    } else if (result === 'no_stock') {
      showNotification('Товар закончился', 'error')
    }
  }

  return (
    <div className={styles.page}>
      {/* Left — market */}
      <div className={styles.left}>
        <div className={styles.pageHeader}>
          <Text variant="h3">Барахолка</Text>
          <Text variant="caption" color="muted">Маркетплейс зоны. Часть товаров — только для своих.</Text>
        </div>

        {/* Category filters */}
        <div className={styles.filters}>
          {MARKET_CATEGORIES.map(cat => (
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

        {/* Items grid */}
        {loading ? (
          <div className={styles.grid}>
            {[1,2,3,4,5,6].map(i => <div key={i} className={styles.skeleton} />)}
          </div>
        ) : (
          <div className={styles.grid}>
            {filtered.map(item => {
              const accessible = hasRankAccess(respect, item.requiredRank)
              const boughtCount = purchased[item.id] ?? 0
              const outOfStock = item.stock !== -1 && boughtCount >= item.stock
              const canBuy = accessible && !outOfStock && balance >= item.price

              return (
                <MarketCard
                  key={item.id}
                  item={item}
                  accessible={accessible}
                  outOfStock={outOfStock}
                  canBuy={canBuy}
                  onBuy={() => handleBuy(item)}
                  requiredRankTitle={
                    item.requiredRank
                      ? RANKS.find(r => r.id === item.requiredRank)?.title
                      : null
                  }
                />
              )
            })}
          </div>
        )}
      </div>

      {/* Right — rank sidebar */}
      <div className={styles.right}>
        {/* Current rank */}
        <Card padding="lg">
          <Text variant="overline" color="muted">Твоё звание</Text>
          <div className={styles.rankDisplay}>
            <span className={styles.rankEmoji}>{currentRank.emoji}</span>
            <div>
              <Text variant="h3" as="span" style={{ color: currentRank.color }}>
                {currentRank.title}
              </Text>
              <Text variant="caption" color="muted" as="div">
                {respect} респекта
              </Text>
            </div>
          </div>

          {nextRank && (
            <div className={styles.rankProgress}>
              <div className={styles.rankProgressLabels}>
                <Text variant="caption" color="muted" as="span">До: {nextRank.emoji} {nextRank.title}</Text>
                <Text variant="caption" color="muted" as="span">{rankProgress}%</Text>
              </div>
              <ProgressBar value={rankProgress} theme="default" size="sm" />
              <Text variant="caption" color="muted">
                Нужно ещё {nextRank.minRespect - respect} респекта
              </Text>
            </div>
          )}
          {!nextRank && (
            <Text variant="caption" color="accent">Максимальное звание достигнуто 👑</Text>
          )}
        </Card>

        {/* All ranks */}
        <Card padding="md">
          <Text variant="label">Все звания</Text>
          <div className={styles.ranksList}>
            {RANKS.map(rank => {
              const unlocked = respect >= rank.minRespect
              return (
                <div
                  key={rank.id}
                  className={[styles.rankRow, !unlocked ? styles.rankLocked : ''].filter(Boolean).join(' ')}
                >
                  <span className={styles.rankRowEmoji}>{rank.emoji}</span>
                  <div className={styles.rankRowInfo}>
                    <Text
                      variant="label"
                      as="span"
                      style={{ color: unlocked ? rank.color : 'var(--color-text-dim)' }}
                    >
                      {rank.title}
                    </Text>
                    <Text variant="caption" color="dim" as="span">
                      от {rank.minRespect} респекта
                    </Text>
                  </div>
                  {unlocked && (
                    <Chip variant="success" size="sm">✓</Chip>
                  )}
                  {!unlocked && (
                    <span className={styles.rankLockIcon}>🔒</span>
                  )}
                </div>
              )
            })}
          </div>
        </Card>
      </div>

      {/* Notification toast */}
      {notification && (
        <div className={[styles.toast, notification.type === 'error' ? styles.toastError : styles.toastSuccess].join(' ')}>
          <Text variant="label" as="span">{notification.msg}</Text>
        </div>
      )}
    </div>
  )
}

function MarketCard({ item, accessible, outOfStock, canBuy, onBuy, requiredRankTitle }) {
  const rank = item.requiredRank ? RANKS.find(r => r.id === item.requiredRank) : null

  return (
    <Card
      padding="md"
      className={[styles.itemCard, !accessible ? styles.itemLocked : ''].filter(Boolean).join(' ')}
    >
      {/* Lock overlay */}
      {!accessible && (
        <div className={styles.lockOverlay}>
          <span className={styles.lockIcon}>🔒</span>
          <Text variant="caption" color="muted" as="span">
            Нужно: {rank?.emoji} {requiredRankTitle}
          </Text>
        </div>
      )}

      {/* Rank badge */}
      {rank && (
        <div className={styles.itemRankBadge}>
          <Chip
            variant={accessible ? 'success' : 'default'}
            size="sm"
          >
            {rank.emoji} {requiredRankTitle}
          </Chip>
        </div>
      )}

      <div className={styles.itemEmoji}>{item.emoji}</div>
      <Text variant="label">{item.title}</Text>
      <Text variant="caption" color="muted">{item.description}</Text>

      <div className={styles.itemFooter}>
        <div className={styles.itemPrice}>
          <Text variant="label" color="accent" as="span">{item.price}</Text>
          <Text variant="caption" color="muted" as="span"> лаве</Text>
        </div>
        {item.stock !== -1 && (
          <Text variant="caption" color={outOfStock ? 'danger' : 'muted'} as="span">
            {outOfStock ? 'нет' : `${item.stock} шт`}
          </Text>
        )}
      </div>

      <Button
        variant={canBuy ? 'secondary' : 'ghost'}
        size="sm"
        fullWidth
        disabled={!accessible || outOfStock}
        onClick={onBuy}
      >
        {outOfStock ? 'Нет в наличии' : !accessible ? 'Закрыто' : 'Купить'}
      </Button>
    </Card>
  )
}
