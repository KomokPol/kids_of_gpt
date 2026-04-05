import { useState } from 'react'
import { Text, Button, Input, Card, ProgressBar, Chip, StarBar, Currency } from '../../ds/index.js'
import styles from './KitchenSinkPage.module.css'

export default function KitchenSinkPage() {
  const [inputVal, setInputVal] = useState('')

  return (
    <div className={styles.page}>
      <header className={styles.header}>
        <Text variant="h2">Design System Kitchen Sink</Text>
        <Text variant="caption" color="muted">Витрина всех компонентов дизайн-системы</Text>
      </header>

      {/* ===== TEXT ===== */}
      <Section title="Text">
        <div className={styles.stack}>
          <Text variant="h1">Заголовок H1</Text>
          <Text variant="h2">Заголовок H2</Text>
          <Text variant="h3">Заголовок H3</Text>
          <Text variant="body">Обычный текст body — описание фильма или контента</Text>
          <Text variant="caption" color="muted">Caption — подпись, год, режиссёр</Text>
          <Text variant="label">Label — метка поля</Text>
          <Text variant="overline">Overline — категория</Text>
        </div>
        <div className={styles.row}>
          <Text color="default">default</Text>
          <Text color="muted">muted</Text>
          <Text color="dim">dim</Text>
          <Text color="accent">accent</Text>
          <Text color="success">success</Text>
          <Text color="warn">warn</Text>
          <Text color="danger">danger</Text>
        </div>
      </Section>

      {/* ===== BUTTON ===== */}
      <Section title="Button">
        <div className={styles.row}>
          <Button variant="primary" size="sm">Primary SM</Button>
          <Button variant="primary" size="md">Primary MD</Button>
          <Button variant="primary" size="lg">Primary LG</Button>
        </div>
        <div className={styles.row}>
          <Button variant="secondary" size="sm">Secondary SM</Button>
          <Button variant="secondary" size="md">Secondary MD</Button>
          <Button variant="secondary" size="lg">Secondary LG</Button>
        </div>
        <div className={styles.row}>
          <Button variant="ghost" size="sm">Ghost SM</Button>
          <Button variant="ghost" size="md">Ghost MD</Button>
          <Button variant="ghost" size="lg">Ghost LG</Button>
        </div>
        <div className={styles.row}>
          <Button variant="primary" disabled>Disabled</Button>
          <Button variant="secondary" disabled>Disabled</Button>
          <Button variant="ghost" disabled>Disabled</Button>
        </div>
        <Button variant="primary" fullWidth>Full Width Button</Button>
        <Button variant="secondary" fullWidth>Full Width Secondary</Button>
      </Section>

      {/* ===== INPUT ===== */}
      <Section title="Input">
        <div className={styles.grid2}>
          <Input
            label="Поиск фильма"
            placeholder="Введите название..."
            value={inputVal}
            onChange={e => setInputVal(e.target.value)}
          />
          <Input
            label="Email"
            placeholder="user@example.com"
            state="default"
          />
          <Input
            label="Ошибка"
            placeholder="Неверный формат"
            state="error"
            hint="Введите корректный email"
          />
          <Input
            label="Успех"
            placeholder="Всё верно"
            state="success"
            hint="Email подтверждён"
          />
          <Input
            label="Отключено"
            placeholder="Недоступно"
            disabled
          />
          <Input
            placeholder="Без лейбла"
          />
        </div>
      </Section>

      {/* ===== CARD ===== */}
      <Section title="Card">
        <div className={styles.grid3}>
          <Card padding="md">
            <Text variant="label">Обычная карточка</Text>
            <Text variant="caption" color="muted">padding=md, без рамки</Text>
          </Card>
          <Card padding="md" highlighted>
            <Text variant="label" color="accent">Highlighted</Text>
            <Text variant="caption" color="muted">Жёлтая рамка — активный элемент</Text>
          </Card>
          <Card padding="md" hoverable>
            <Text variant="label">Hoverable</Text>
            <Text variant="caption" color="muted">Наведи мышь</Text>
          </Card>
          <Card padding="sm">
            <Text variant="caption">padding=sm</Text>
          </Card>
          <Card padding="lg">
            <Text variant="label">padding=lg</Text>
            <Text variant="caption" color="muted">Больше отступов</Text>
          </Card>
          <Card padding="md" highlighted hoverable>
            <Text variant="label" color="accent">Highlighted + Hoverable</Text>
          </Card>
        </div>
      </Section>

      {/* ===== PROGRESS BAR ===== */}
      <Section title="ProgressBar">
        <div className={styles.stack}>
          <div>
            <Text variant="caption" color="muted">default — 65%</Text>
            <ProgressBar value={65} theme="default" />
          </div>
          <div>
            <Text variant="caption" color="muted">warn — 40%</Text>
            <ProgressBar value={40} theme="warn" />
          </div>
          <div>
            <Text variant="caption" color="muted">danger — 15%</Text>
            <ProgressBar value={15} theme="danger" />
          </div>
          <div>
            <Text variant="caption" color="muted">success — 100%</Text>
            <ProgressBar value={100} theme="success" />
          </div>
          <div>
            <Text variant="caption" color="muted">size=sm</Text>
            <ProgressBar value={55} size="sm" />
          </div>
          <div>
            <Text variant="caption" color="muted">size=lg + showLabel</Text>
            <ProgressBar value={78} size="lg" showLabel />
          </div>
        </div>
      </Section>

      {/* ===== CHIP ===== */}
      <Section title="Chip">
        <div className={styles.row}>
          <Chip variant="default">default</Chip>
          <Chip variant="warn">warn</Chip>
          <Chip variant="danger">danger</Chip>
          <Chip variant="success">success</Chip>
          <Chip variant="info">info</Chip>
        </div>
        <div className={styles.row}>
          <Chip variant="default" size="sm">sm default</Chip>
          <Chip variant="warn" size="sm">sm warn</Chip>
          <Chip variant="danger" size="sm">sm danger</Chip>
          <Chip variant="success" size="sm">sm success</Chip>
        </div>
        <div className={styles.row}>
          <Text variant="caption" color="muted">Фильтры:</Text>
          <Chip variant="default" active>Всё</Chip>
          <Chip variant="default">Тюрьма</Chip>
          <Chip variant="default">90-е</Chip>
          <Chip variant="default">Деревня</Chip>
          <Chip variant="default">Балабанов</Chip>
          <Chip variant="warn">ХИТ</Chip>
        </div>
      </Section>

      {/* ===== STAR BAR ===== */}
      <Section title="StarBar">
        <div className={styles.stack}>
          <div className={styles.row}>
            <Text variant="caption" color="muted" style={{ width: 80 }}>5 из 5:</Text>
            <StarBar value={5} max={5} showValue />
          </div>
          <div className={styles.row}>
            <Text variant="caption" color="muted" style={{ width: 80 }}>3.7 из 5:</Text>
            <StarBar value={3.7} max={5} showValue />
          </div>
          <div className={styles.row}>
            <Text variant="caption" color="muted" style={{ width: 80 }}>1.5 из 5:</Text>
            <StarBar value={1.5} max={5} showValue />
          </div>
          <div className={styles.row}>
            <Text variant="caption" color="muted" style={{ width: 80 }}>8.1 из 10:</Text>
            <StarBar value={8.1} max={10} showValue />
          </div>
          <div className={styles.row}>
            <Text variant="caption" color="muted" style={{ width: 80 }}>0 из 5:</Text>
            <StarBar value={0} max={5} showValue />
          </div>
          <div className={styles.row}>
            <Text variant="caption" color="muted" style={{ width: 80 }}>2.3 из 5:</Text>
            <StarBar value={2.3} max={5} showValue />
          </div>
        </div>
      </Section>

      {/* ===== CURRENCY ===== */}
      <Section title="Currency">
        <div className={styles.row}>
          <Currency amount={290} />
          <Currency amount={890} unit="лаве/мес" size="lg" />
          <Currency amount={2400} unit="лаве/мес" size="xl" color="accent" />
          <Currency amount={0} unit="лаве/мес" size="lg" />
          <Currency free size="md" />
          <Currency free size="lg" />
        </div>
        <div className={styles.row}>
          <Currency amount={490} size="sm" color="muted" />
          <Currency amount={490} size="md" />
          <Currency amount={490} size="lg" color="accent" />
          <Currency amount={490} size="xl" />
        </div>
      </Section>

      {/* ===== COMBINATIONS ===== */}
      <Section title="Комбинации компонентов">
        <div className={styles.grid3}>
          {/* Карточка фильма */}
          <Card padding="md" hoverable>
            <div className={styles.movieCard}>
              <div className={styles.moviePoster}>🎬</div>
              <div className={styles.movieInfo}>
                <Text variant="label">Груз 200</Text>
                <Text variant="caption" color="muted">Балабанов · 2007</Text>
                <div className={styles.row}>
                  <StarBar value={9.1} max={10} showValue />
                  <Currency amount={490} size="sm" />
                </div>
              </div>
            </div>
          </Card>

          {/* Карточка тарифа */}
          <Card padding="lg" highlighted>
            <div className={styles.stack}>
              <div className={styles.row}>
                <Text variant="h3">Пацан</Text>
                <Chip variant="warn" size="sm">ХИТ</Chip>
              </div>
              <Currency amount={890} unit="лаве/мес" size="lg" color="accent" />
              <Text variant="caption" color="muted">Весь каталог. Без рекламы. 1080p.</Text>
              <Button variant="primary" fullWidth>Оформить</Button>
            </div>
          </Card>

          {/* Карточка прогресса */}
          <Card padding="md">
            <div className={styles.stack}>
              <Text variant="label">Прогресс уровня</Text>
              <div className={styles.row}>
                <Text variant="caption" color="muted">Ур. 5</Text>
                <Chip variant="success" size="sm">XP 575</Chip>
              </div>
              <ProgressBar value={72} theme="default" size="md" />
              <Text variant="caption" color="muted">До следующего уровня: 28%</Text>
            </div>
          </Card>

          {/* Форма отправки */}
          <Card padding="md">
            <div className={styles.stack}>
              <Text variant="label">Подогреть другого</Text>
              <Input placeholder="Никнейм получателя" />
              <Input placeholder="Записка (доступна с ур. 2)" disabled />
              <Button variant="secondary" fullWidth>Подогреть анонимно</Button>
            </div>
          </Card>

          {/* Статусы */}
          <Card padding="md">
            <div className={styles.stack}>
              <Text variant="label">Статусы</Text>
              <div className={styles.row}>
                <Chip variant="success">Доставлено</Chip>
                <Chip variant="warn">В пути</Chip>
                <Chip variant="danger">Отменено</Chip>
              </div>
              <div className={styles.row}>
                <Chip variant="info">Новинка</Chip>
                <Chip variant="default">Архив</Chip>
                <Chip variant="warn">ХИТ</Chip>
              </div>
            </div>
          </Card>

          {/* Рейтинги */}
          <Card padding="md">
            <div className={styles.stack}>
              <Text variant="label">Рейтинги фильмов</Text>
              {[
                { title: 'Брат', rating: 8.4, price: 350 },
                { title: 'Бумер', rating: 8.2, price: 290 },
                { title: 'Жмурки', rating: 8.0, price: 240 },
                { title: 'Левиафан', rating: 7.7, price: 350 },
              ].map(film => (
                <div key={film.title} className={styles.filmRow}>
                  <Text variant="caption">{film.title}</Text>
                  <StarBar value={film.rating} max={10} showValue />
                  <Currency amount={film.price} size="sm" color="muted" />
                </div>
              ))}
            </div>
          </Card>
        </div>
      </Section>
    </div>
  )
}

function Section({ title, children }) {
  return (
    <section className={styles.section}>
      <Text variant="overline" color="muted">{title}</Text>
      <div className={styles.sectionContent}>{children}</div>
    </section>
  )
}
