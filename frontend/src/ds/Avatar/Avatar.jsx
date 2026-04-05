import { useState } from 'react'
import styles from './Avatar.module.css'

/**
 * Avatar — иконка пользователя
 *
 * @param {string} src — URL изображения
 * @param {string} login — логин пользователя (для fallback инициалов)
 * @param {'sm'|'md'|'lg'} size
 * @param {string} className
 */
export default function Avatar({
  src,
  login = '',
  size = 'md',
  className = '',
  ...props
}) {
  const [imgError, setImgError] = useState(false)

  // Берём первые 2 символа логина как инициалы
  const initials = login
    ? login.slice(0, 2).toUpperCase()
    : '?'

  const showImage = src && !imgError

  return (
    <span
      className={[styles.avatar, styles[size], className].filter(Boolean).join(' ')}
      title={login}
      {...props}
    >
      {showImage ? (
        <img
          src={src}
          alt={login}
          className={styles.img}
          onError={() => setImgError(true)}
        />
      ) : (
        <span className={styles.initials}>{initials}</span>
      )}
    </span>
  )
}
