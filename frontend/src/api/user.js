/**
 * API-слой для пользователя
 *
 * Сейчас — заглушка.
 * TODO: когда будет готов бэкенд, заменить тело функции на:
 *
 *   const res = await fetch('/api/user')
 *   if (!res.ok) throw new Error(`HTTP ${res.status}`)
 *   return res.json()
 */
export async function getUser() {
  // Имитация сетевой задержки
  await new Promise(r => setTimeout(r, 300))

  // Заглушка — мок-данные пользователя
  return {
    login:   'Балабанов',
    iconUrl: null,      // null → Avatar покажет инициалы
    respect: 575,
    balance: 1240,
  }
}
