export function safeParse(key) {
  try {
    const val = localStorage.getItem(key)
    if (!val || val === 'undefined' || val === 'null') return null
    return JSON.parse(val)
  } catch {
    localStorage.removeItem(key)
    return null
  }
}