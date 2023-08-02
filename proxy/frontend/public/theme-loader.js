const THEME_STORAGE_KEY = "color-theme";
const DARK_THEME = "dark";
const DEFAULT_THEME = "default";

const prefersDark = () => window.matchMedia('(prefers-color-scheme: dark)').matches;

const applyColorTheme = (theme) => {
  document.documentElement.setAttribute(THEME_STORAGE_KEY, theme)
  localStorage.setItem(THEME_STORAGE_KEY, theme)
}

const getCurrentAppliedColorTheme = () => {
  const currentAttribute = document.documentElement.getAttribute(THEME_STORAGE_KEY);
  if(currentAttribute)
    return currentAttribute;

  return DEFAULT_THEME;
}

const toggleDarkTheme = () => {
  if(getCurrentAppliedColorTheme() === DARK_THEME) {
    applyColorTheme(DEFAULT_THEME)
    return false;
  }
  else {
    applyColorTheme(DARK_THEME)
    return true;
  }
}

const storedTheme = localStorage.getItem(THEME_STORAGE_KEY);
if(storedTheme)
  applyColorTheme(storedTheme)
else if(prefersDark())
  applyColorTheme(DARK_THEME)

