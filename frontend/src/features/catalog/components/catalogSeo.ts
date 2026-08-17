export function catalogRobots(noindex: boolean, search: string) {
  return noindex || search !== '' ? 'noindex, follow' : 'index, follow'
}
