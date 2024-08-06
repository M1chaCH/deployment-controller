// place files you want to import through the `$lib` alias in this folder.

export function registerCloseBackdrop(close: () => {}, exactHit: boolean) {
  document.getElementsByClassName("backdrop")[0]?.addEventListener("click", (e) => {
    if(exactHit) {
      const element = e.target as HTMLElement;
      if(element.classList.contains("backdrop")) {
        close()
      }
    } else {
      close();
    }
  })
}