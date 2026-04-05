import { useRef } from 'react'
import styles from './Input.module.css'

/**
 * Input — текстовое поле дизайн-системы
 *
 * @param {string} label — подпись над полем
 * @param {string} hint — подсказка под полем
 * @param {string} placeholder
 * @param {boolean} disabled
 * @param {'default'|'error'|'success'} state
 * @param {React.ReactNode} leftIcon
 * @param {React.ReactNode} rightIcon
 */
export default function Input({
  label,
  hint,
  placeholder,
  disabled = false,
  state = 'default',
  leftIcon,
  rightIcon,
  className = '',
  id,
  ...props
}) {
  const inputRef = useRef(null)
  const inputId = id || (label ? label.toLowerCase().replace(/\s+/g, '-') : undefined)

  const handleWrapClick = () => {
    if (!disabled) inputRef.current?.focus()
  }

  return (
    <div className={[styles.wrapper, className].filter(Boolean).join(' ')}>
      {label && (
        <label htmlFor={inputId} className={styles.label}>
          {label}
        </label>
      )}
      <div
        className={[styles.inputWrap, styles[state], disabled ? styles.disabled : ''].filter(Boolean).join(' ')}
        onClick={handleWrapClick}
      >
        {leftIcon && <span className={styles.icon}>{leftIcon}</span>}
        <input
          ref={inputRef}
          id={inputId}
          className={styles.input}
          placeholder={placeholder}
          disabled={disabled}
          {...props}
        />
        {rightIcon && <span className={styles.icon}>{rightIcon}</span>}
      </div>
      {hint && (
        <span className={[styles.hint, state === 'error' ? styles.hintError : ''].filter(Boolean).join(' ')}>
          {hint}
        </span>
      )}
    </div>
  )
}
