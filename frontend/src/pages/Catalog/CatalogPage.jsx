import { useState } from 'react'
import { Text, Button, Card, Chip, StarBar, Currency } from '../../ds/index.js'
import styles from './CatalogPage.module.css'

const FILTERS = ['Всё', 'Тюрьма', '90-е', 'Деревня', 'Балабанов', 'Безнадёга', 'Про войну']

const FILMS = [
  { id: 1, title: 'Груз 200',         year: 1984, director: 'Балабанов', rating: 9.1, price: 490,  free: false, tags: ['Тюрьма', 'Деревня', 'Балабанов'] },
  { id: 2, title: 'Брат',             year: 1997, director: 'Балабанов', rating: 8.4, price: null, free: true,  tags: ['90-е', 'Балабанов'] },
  { id: 3, title: 'Бумер',            year: 2003, director: 'Буслов',    rating: 8.2, price: 290,  free: false, tags: ['90-е', 'Тюрьма'] },
  { id: 4, title: 'Жмурки',           year: 2005, director: 'Михалков',  rating: 8.0, price: 240,  free: false, tags: ['90-е'] },
  { id: 5, title: 'Левиафан',         year: 2014, director: 'Звягинцев', rating: 7.7, price: 350,  free: false, tags: ['Деревня', 'Безнадёга'] },
  { id: 6, title: 'Дурак',            year: 2014, director: 'Быков',     rating: 8.1, price: 320,  free: false, tags: ['Безнадёга'] },
  { id: 7, title: 'Кококо',           year: 2012, director: 'Смирнова',  rating: 7.2, price: 180,  free: false, tags: ['90-е'] },
  { id: 8, title: 'Про уродов и людей', year: 1998, director: 'Балабанов', rating: 8.3, price: 410, free: false, tags: ['Балабанов', 'Безнадёга'] },
  { id: 9, title: 'Сынок',            year: 2009, director: 'Карасёв',   rating: 7.5, price: 270,  free: false, tags: ['Тюрьма'] },
  { id: 10, title: 'Зелёный слоник',  year: 1999, director: 'Лихачёва',  rating: 6.8, price: 150,  free: false, tags: ['Тюрьма', 'Безнадёга'] },
  { id: 11, title: 'Брат 2',          year: 2000, director: 'Балабанов', rating: 8.0, price: 320,  free: false, tags: ['90-е', 'Балабанов'] },
  { id: 12, title: 'Война',           year: 2002, director: 'Балабанов', rating: 7.6, price: 280,  free: false, tags: ['Балабанов', 'Про войну'] },
]

export default function CatalogPage() {
  const [activeFilter, setActiveFilter] = useState('Всё')

  const filtered = activeFilter === 'Всё'
    ? FILMS
    : FILMS.filter(f => f.tags.includes(activeFilter))

  return (
    <div className={styles.page}>
      {/* Hero */}
      <section className={styles.hero}>
        <div className={styles.heroLeft}>
          <Text variant="overline" color="muted">Фильм недели</Text>
          <Text variant="h1">Груз <Text variant="h1" as="span" color="accent">200</Text></Text>
          <Text variant="body" color="muted">
            1984 год. Советская провинция. Режиссёр Балабанов снял то, что многие хотели бы забыть.
            Самый тяжёлый фильм в истории российского кино.
          </Text>
          <div className={styles.heroActions}>
            <Button variant="primary" size="lg">Смотреть — 490 лаве</Button>
            <Button variant="secondary" size="lg">🔒 Бесплатно с подпиской</Button>
          </div>
        </div>
        <Card padding="md" className={styles.heroCard}>
          <div className={styles.heroCardInner}>
            <div className={styles.heroRating}>
              <Text variant="label" color="accent">9.1</Text>
            </div>
            <div className={styles.heroPoster}>🎬</div>
            <Text variant="caption" color="muted">Балабанов · 2007</Text>
          </div>
        </Card>
      </section>

      {/* Catalog section */}
      <section className={styles.catalogSection}>
        <div className={styles.catalogHeader}>
          <Text variant="h3">Хтонь и безысходность</Text>
          <Text variant="caption" color="muted">{FILMS.length} фильмов</Text>
        </div>

        {/* Filters */}
        <div className={styles.filters}>
          {FILTERS.map(f => (
            <Chip
              key={f}
              variant="default"
              active={f === activeFilter}
              style={{ cursor: 'pointer' }}
              onClick={() => setActiveFilter(f)}
            >
              {f}
            </Chip>
          ))}
        </div>

        {/* Film grid */}
        <div className={styles.filmGrid}>
          {filtered.map(film => (
            <FilmCard key={film.id} film={film} />
          ))}
        </div>
      </section>
    </div>
  )
}

function FilmCard({ film }) {
  return (
    <Card padding="sm" hoverable className={styles.filmCard}>
      <div className={styles.filmPoster}>
        <span className={styles.filmEmoji}>🔒</span>
      </div>
      <div className={styles.filmMeta}>
        <Text variant="label">{film.title}</Text>
        <Text variant="caption" color="muted">{film.year}</Text>
        <div className={styles.filmBottom}>
          <StarBar value={film.rating} max={10} showValue />
          {film.free
            ? <Currency free size="sm" />
            : <Currency amount={film.price} size="sm" color="muted" />
          }
        </div>
      </div>
    </Card>
  )
}
