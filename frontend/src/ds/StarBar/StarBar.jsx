import styles from './StarBar.module.css'

/**
 * StarBar — рейтинг в виде звёздочек
 *
 * @param {number} value — значение рейтинга (0–max, может быть нецелым: 3.7)
 * @param {number} max — максимум звёзд (default: 5)
 * @param {boolean} showValue — показывать числовое значение рядом
 */
export default function StarBar({
  value = 0,
  max = 5,
  showValue = false,
  className = '',
  ...props
}) {
  const clamped = Math.min(max, Math.max(0, value))
  const percentage = (clamped / max) * 100

  return (
    <span
      className={[styles.starBar, className].filter(Boolean).join(' ')}
      title={`${clamped} из ${max}`}
      {...props}
    >
      <span className={styles.starsTrack} aria-hidden="true">
        {/* Bottom layer: empty (grey) stars */}
        <span className={styles.starsEmpty}>
          {Array.from({ length: max }).map((_, i) => (
            <StarIcon key={i} />
          ))}
        </span>
        {/* Top layer: filled (yellow) stars, clipped by width */}
        <span
          className={styles.starsFilled}
          style={{ width: `${percentage}%` }}
        >
          {Array.from({ length: max }).map((_, i) => (
            <StarIcon key={i} filled />
          ))}
        </span>
      </span>
      {showValue && (
        <span className={styles.value}>
          {clamped % 1 === 0 ? clamped.toFixed(1) : clamped}
        </span>
      )}
    </span>
  )
}

function StarIcon({ filled = false }) {
  return (
    <svg
      className={styles.starIcon}
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
    >
      {filled ? (
        <path
          fill="currentColor"
          d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
        />
      ) : (
        <path
          fill="none"
          stroke="currentColor"
          strokeWidth="1.5"
          strokeLinejoin="round"
          d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"
        />
      )}
    </svg>
  )
}
