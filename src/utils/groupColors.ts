export const groupColors = ['blue', 'indigo', 'purple', 'pink', 'orange', 'mint', 'brown', 'teal', 'green', 'red'] as const
export const groupTints = ['#E6F1FC', '#ECECF8', '#F4ECF7', '#FCE8EE', '#FBEBE5', '#E4F3F2', '#F3EFEA', '#E5F1F6', '#E8F4EB', '#FCE6E8'] as const
export const groupForegrounds = ['#0062C4', '#3634A3', '#8944AB', '#C4103F', '#B83000', '#0C7A74', '#7F6545', '#0070A0', '#1B7330', '#C40012'] as const

export function iconForeground(background?: string) {
  const index = groupTints.findIndex(tint => tint.toLowerCase() === background?.toLowerCase())
  return index >= 0 ? groupForegrounds[index] : undefined
}
