import styles from './Currency.module.css'

/**
 * Currency — отображение суммы/цены
 *
 * @param {number|string} amount — сумма (число или строка)
 * @param {string} unit — единица измерения (default: "лаве")
 * @param {boolean} free — бесплатно (перекрывает amount)
 * @param {'sm'|'md'|'lg'|'xl'} size
 * @param {'default'|'accent'|'muted'} color
 */
export default function Currency({
  amount,
  unit = 'лаве',
  free = false,
  size = 'md',
  color = 'default',
  className = '',
  ...props
}) {
  if (free) {
    return (
      <span
        className={[styles.currency, styles[size], styles.free, className]
          .filter(Boolean)
          .join(' ')}
        {...props}
      >
        бесплатно
      </span>
    )
  }

  return (
    <span
      className={[styles.currency, styles[size], styles[`color_${color}`], className]
        .filter(Boolean)
        .join(' ')}
      {...props}
    >
      <span className={styles.amount}>{amount}</span>
      {unit && <span className={styles.unit}>{unit}</span>}
    </span>
  )
}
