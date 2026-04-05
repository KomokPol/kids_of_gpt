import styles from './ProgressBar.module.css'

/**
 * ProgressBar — полоса прогресса
 *
 * @param {number} value — заполненность 0–100
 * @param {'default'|'warn'|'danger'|'success'} theme
 * @param {'sm'|'md'|'lg'} size
 * @param {boolean} showLabel — показывать процент
 */
export default function ProgressBar({
  value = 0,
  theme = 'default',
  size = 'md',
  showLabel = false,
  className = '',
  ...props
}) {
  const clamped = Math.min(100, Math.max(0, value))

  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')} {...props}>
      {showLabel && (
        <span className={styles.labelText}>{clamped}%</span>
      )}
      <div className={[styles.track, styles[size]].join(' ')}>
        <div
          className={[styles.fill, styles[theme]].join(' ')}
          style={{ width: `${clamped}%` }}
          role="progressbar"
          aria-valuenow={clamped}
          aria-valuemin={0}
          aria-valuemax={100}
        />
      </div>
    </div>
  )
}
