import styles from './Card.module.css'

/**
 * Card — тёмная карточка дизайн-системы
 *
 * @param {boolean} highlighted — жёлтая рамка (активный/выбранный)
 * @param {boolean} hoverable — эффект при наведении
 * @param {'sm'|'md'|'lg'} padding
 * @param {string} className
 */
export default function Card({
  highlighted = false,
  hoverable = false,
  padding = 'md',
  className = '',
  children,
  ...props
}) {
  return (
    <div
      className={[
        styles.card,
        styles[`padding_${padding}`],
        highlighted ? styles.highlighted : '',
        hoverable ? styles.hoverable : '',
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      {...props}
    >
      {children}
    </div>
  )
}
