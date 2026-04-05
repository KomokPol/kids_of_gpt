import styles from './Chip.module.css'

/**
 * Chip — компонент статуса / тега
 *
 * @param {'default'|'warn'|'danger'|'success'|'info'} variant
 * @param {'sm'|'md'} size
 * @param {React.ReactNode} icon
 * @param {boolean} active — выделенный/активный чип (например, выбранный фильтр)
 */
export default function Chip({
  variant = 'default',
  size = 'md',
  icon,
  active = false,
  className = '',
  children,
  ...props
}) {
  return (
    <span
      className={[
        styles.chip,
        styles[variant],
        styles[size],
        active ? styles.active : '',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      {...props}
    >
      {icon && <span className={styles.icon}>{icon}</span>}
      <span className={styles.label}>{children}</span>
    </span>
  )
}
