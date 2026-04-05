import { NavLink } from 'react-router-dom'
import { Text, Avatar, Currency } from '../../ds/index.js'
import useUserStore from '../../store/useUserStore.js'
import styles from './Header.module.css'

/**
 * Header — глобальный хедер приложения Zondax
 * Продуктовый компонент в src/ui/, собран из DS-примитивов
 */
export default function Header() {
  const { login, iconUrl, respect, balance, loading } = useUserStore()

  return (
    <header className={styles.header}>
      <div className={styles.inner}>
        {/* Logo */}
        <NavLink to="/" className={styles.logo}>
          <Text variant="h3" as="span" className={styles.logoText}>Zondax</Text>
        </NavLink>

        {/* Nav */}
        <nav className={styles.nav}>
          <NavLink
            to="/"
            end
            className={({ isActive }) =>
              [styles.navLink, isActive ? styles.navLinkActive : ''].filter(Boolean).join(' ')
            }
          >
            <Text variant="label" as="span">Kitchen Sink</Text>
          </NavLink>
          <NavLink
            to="/games"
            className={({ isActive }) =>
              [styles.navLink, isActive ? styles.navLinkActive : ''].filter(Boolean).join(' ')
            }
          >
            <Text variant="label" as="span">Игры</Text>
          </NavLink>
          <NavLink
            to="/edu"
            className={({ isActive }) =>
              [styles.navLink, isActive ? styles.navLinkActive : ''].filter(Boolean).join(' ')
            }
          >
            <Text variant="label" as="span">Шарага</Text>
          </NavLink>
          <NavLink
            to="/catalog"
            className={({ isActive }) =>
              [styles.navLink, isActive ? styles.navLinkActive : ''].filter(Boolean).join(' ')
            }
          >
            <Text variant="label" as="span">Каталог</Text>
          </NavLink>
        </nav>

        {/* User info */}
        <div className={styles.user}>
          {loading ? (
            <span className={styles.userLoading} />
          ) : (
            <>
              <div className={styles.userStats}>
                <span className={styles.stat}>
                  <Text variant="label" as="span" style={{ color: 'var(--color-respect)' }}>{respect}</Text>
                  <Text variant="label" color="muted" as="span">респект</Text>
                </span>
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
