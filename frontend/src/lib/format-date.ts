export const formatDate = (iso: string) => new Intl.DateTimeFormat('ru-RU', {
  dateStyle: 'medium',
  timeStyle: 'short'
}).format(new Date(iso));
