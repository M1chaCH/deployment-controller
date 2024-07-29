export type MichuTechColorTheme = "default" | "dark";
export type ThemeChangeHandler = (theme: MichuTechColorTheme) => void;

const THEME_STORAGE_KEY = "color-theme";
const changeHandlers = new Map<number, ThemeChangeHandler>
let handlerIndex = 0;

export const prefersDark = () => window.matchMedia('(prefers-color-scheme: dark)').matches;

export const applyColorTheme = (theme: MichuTechColorTheme) => {
  document.documentElement.setAttribute(THEME_STORAGE_KEY, theme)
  localStorage.setItem(THEME_STORAGE_KEY, theme)
  notifyHandlers(theme)
}

export const getCurrentAppliedColorTheme = (): MichuTechColorTheme => {
  const currentAttribute = document.documentElement.getAttribute(THEME_STORAGE_KEY);
  if(currentAttribute)
    return currentAttribute.toLocaleLowerCase() as MichuTechColorTheme;

  return "default";
}

export const toggleDarkTheme = () => {
  if(getCurrentAppliedColorTheme() === "dark") {
    applyColorTheme("default")
  }
  else {
    applyColorTheme("dark")
  }
}

export const initColorTheme = () => {
  const storedTheme = localStorage.getItem(THEME_STORAGE_KEY) as MichuTechColorTheme | null;
  if(storedTheme)
    applyColorTheme(storedTheme)
  else if(prefersDark())
    applyColorTheme("dark")
}

export const registerThemeChangeHandler = (handler: ThemeChangeHandler): number => {
  handlerIndex++;
  changeHandlers.set(handlerIndex, handler)
  return handlerIndex
}

export const removeThemeChangeHandler = (index: number) => {
  changeHandlers.delete(index)
}

const notifyHandlers = (theme: MichuTechColorTheme) => {
  for (let value of changeHandlers.values()) {
    value(theme)
  }
}