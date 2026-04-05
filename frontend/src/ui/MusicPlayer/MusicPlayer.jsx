import { useState, useRef, useEffect } from 'react'
import { Text } from '../../ds/index.js'
import styles from './MusicPlayer.module.css'

const PLAYLIST = [
  { id: 1, title: 'Этапирование',    artist: 'Зонный оркестр',  duration: 214 },
  { id: 2, title: 'Баланда блюз',    artist: 'Смотрящий бэнд',  duration: 187 },
  { id: 3, title: 'Пайка хлеба',     artist: 'МаксТопор',       duration: 203 },
  { id: 4, title: 'Общак',           artist: 'КолянЗэк feat. ВаняРулон', duration: 256 },
  { id: 5, title: 'На волю',         artist: 'Балабанов Ремикс', duration: 178 },
]

function formatTime(sec) {
  const m = Math.floor(sec / 60)
  const s = String(Math.floor(sec % 60)).padStart(2, '0')
  return `${m}:${s}`
}

export default function MusicPlayer() {
  const [isOpen, setIsOpen]       = useState(true)
  const [isPlaying, setIsPlaying] = useState(false)
  const [currentIdx, setCurrentIdx] = useState(0)
  const [progress, setProgress]   = useState(0)   // 0–100
  const [elapsed, setElapsed]     = useState(0)   // секунды

  const intervalRef = useRef(null)
  const track = PLAYLIST[currentIdx]

  // Имитация воспроизведения — тикаем таймер
  useEffect(() => {
    if (isPlaying) {
      intervalRef.current = setInterval(() => {
        setElapsed(prev => {
          const next = prev + 1
          if (next >= track.duration) {
            // Переход к следующему треку
            handleNext()
            return 0
          }
          setProgress(Math.round((next / track.duration) * 100))
          return next
        })
      }, 1000)
    } else {
      clearInterval(intervalRef.current)
    }
    return () => clearInterval(intervalRef.current)
  }, [isPlaying, currentIdx])

  const handlePlayPause = () => setIsPlaying(p => !p)

  const handlePrev = () => {
    setCurrentIdx(i => (i - 1 + PLAYLIST.length) % PLAYLIST.length)
    setElapsed(0)
    setProgress(0)
  }

  const handleNext = () => {
    setCurrentIdx(i => (i + 1) % PLAYLIST.length)
    setElapsed(0)
    setProgress(0)
  }

  const handleProgressClick = (e) => {
    const rect = e.currentTarget.getBoundingClientRect()
    const ratio = (e.clientX - rect.left) / rect.width
    const newElapsed = Math.floor(ratio * track.duration)
    setElapsed(newElapsed)
    setProgress(Math.round(ratio * 100))
  }

  if (!isOpen) {
    return (
      <button className={styles.miniBtn} onClick={() => setIsOpen(true)} title="Открыть плеер">
        🎵
      </button>
    )
  }

  return (
    <div className={styles.player}>
      {/* Header */}
      <div className={styles.header}>
        <Text variant="overline" color="muted">Зонный FM</Text>
        <button className={styles.closeBtn} onClick={() => setIsOpen(false)}>✕</button>
      </div>

      {/* Track info */}
      <div className={styles.trackInfo}>
        <div className={styles.albumArt}>
          {isPlaying ? '🎵' : '🎶'}
        </div>
        <div className={styles.trackMeta}>
          <Text variant="label" className={styles.trackTitle}>{track.title}</Text>
          <Text variant="caption" color="muted">{track.artist}</Text>
        </div>
      </div>

      {/* Progress bar */}
      <div className={styles.progressWrap} onClick={handleProgressClick}>
        <div className={styles.progressTrack}>
          <div className={styles.progressFill} style={{ width: `${progress}%` }} />
        </div>
        <div className={styles.timeRow}>
          <Text variant="caption" color="muted" as="span">{formatTime(elapsed)}</Text>
          <Text variant="caption" color="muted" as="span">{formatTime(track.duration)}</Text>
        </div>
      </div>

      {/* Controls */}
      <div className={styles.controls}>
        <button className={styles.ctrlBtn} onClick={handlePrev} title="Предыдущий">⏮</button>
        <button className={[styles.ctrlBtn, styles.playBtn].join(' ')} onClick={handlePlayPause}>
          {isPlaying ? '⏸' : '▶'}
        </button>
        <button className={styles.ctrlBtn} onClick={handleNext} title="Следующий">⏭</button>
      </div>

      {/* Playlist */}
      <div className={styles.playlist}>
        {PLAYLIST.map((t, i) => (
          <div
            key={t.id}
            className={[styles.playlistItem, i === currentIdx ? styles.playlistItemActive : ''].filter(Boolean).join(' ')}
            onClick={() => { setCurrentIdx(i); setElapsed(0); setProgress(0) }}
          >
            <Text variant="caption" color={i === currentIdx ? 'accent' : 'muted'} as="span" className={styles.playlistNum}>
              {i === currentIdx && isPlaying ? '♪' : i + 1}
            </Text>
            <Text variant="caption" as="span" className={styles.playlistTitle}>{t.title}</Text>
            <Text variant="caption" color="muted" as="span">{formatTime(t.duration)}</Text>
          </div>
        ))}
      </div>
    </div>
  )
}
