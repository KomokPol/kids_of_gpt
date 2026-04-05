import { NavLink } from 'react-router-dom'
import { Text, Avatar } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import { getRankByRespect } from '../../config/ranks.js'
import { hasAccess } from '../../config/access.js'
import styles from './Header.module.css'

const NAV_ITEMS = [
  { path: '/games',   label: 'Игры',       sectionId: 'games' },
  { path: '/market',  label: 'Барахолка',  sectionId: 'market' },
  { path: '/edu',     label: 'Шарага',     sectionId: 'edu' },
  { path: '/catalog', label: 'Каталог',    sectionId: 'catalog' },
]

export default function Header() {
  const { login, iconUrl, respect, balance, loading } = useUserStore()
  const rank = getRankByRespect(respect)

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        {/* Logo */}
        <NavLink to="/" className={styles.logo}>
          <Text variant="h3" as="span" className={styles.logoText}>Zondax</Text>
        </NavLink>

        {/* Nav */}
        <nav className={styles.nav}>
          {NAV_ITEMS.map(item => {
            const accessible = hasAccess(item.sectionId, respect)
            return (
              <NavLink
                key={item.path}
                to={item.path}
                className={({ isActive }) =>
                  [
                    styles.navLink,
                    isActive ? styles.navLinkActive : '',
                    !accessible ? styles.navLinkLocked : '',
                  ].filter(Boolean).join(' ')
                }
                title={!accessible ? `Нужен ранг выше` : undefined}
                onClick={e => { if (!accessible) e.preventDefault() }}
              >
                {!accessible && <span className={styles.navLock}>🔒</span>}
                <Text variant="label" as="span">{item.label}</Text>
              </NavLink>
            )
          })}
        </nav>

        {/* User info */}
        <div className={styles.user}>
          {loading ? (
            <span className={styles.userLoading} />
          ) : (
            <>
              <div className={styles.userStats}>
                {/* Звание */}
                <span className={styles.stat} title={`${rank.title} — от ${rank.minRespect} респекта`}>
                  <Text variant="label" as="span">{rank.emoji}</Text>
                  <Text variant="label" as="span" style={{ color: rank.color }}>{rank.title}</Text>
                </span>
                {/* Респект */}
                <span className={styles.stat}>
                  <Text variant="label" as="span" style={{ color: 'var(--color-respect)' }}>{respect}</Text>
                  <Text variant="label" color="muted" as="span">респект</Text>
                </span>
                {/* Баланс */}
                <span className={styles.stat}>
                  <Text variant="label" color="accent" as="span">{balance}</Text>
                  <Text variant="label" color="muted" as="span">лаве</Text>
                </span>
              </div>
              <Avatar src={iconUrl} login={login ?? ''} size="sm" />
              {login && (
                <Text variant="label" as="span" className={styles.userLogin}>
                  {login}
                </Text>
              )}
            </>
          )}
        </div>
      </div>
    </header>
  )
}
