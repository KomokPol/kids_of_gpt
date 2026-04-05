import styles from './Text.module.css'

/**
 * Text — базовый типографический компонент
 *
 * @param {'h1'|'h2'|'h3'|'body'|'caption'|'label'|'overline'} variant
 * @param {'default'|'muted'|'dim'|'accent'|'success'|'warn'|'danger'} color
 * @param {string} as — переопределить HTML-тег
 * @param {string} className
 */
export default function Text({
  variant = 'body',
  color = 'default',
  as,
  className = '',
  children,
  ...props
}) {
  const tagMap = {
    h1: 'h1',
    h2: 'h2',
    h3: 'h3',
    body: 'p',
    caption: 'span',
    label: 'span',
    overline: 'span',
  }

  const Tag = as || tagMap[variant] || 'span'

  return (
    <Tag
      className={[
        styles.text,
        styles[variant],
        styles[`color_${color}`],
        className,
      ]
        .filter(Boolean)
        .join(' ')}
      {...props}
    >
      {children}
    </Tag>
  )
}
