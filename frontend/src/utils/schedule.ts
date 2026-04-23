export function scheduleAfterPaint(task: () => void, delayMs = 120): () => void {
  if (typeof window === "undefined") {
    return () => {};
  }

  let raf1 = 0;
  let raf2 = 0;
  let timer = 0;

  raf1 = window.requestAnimationFrame(() => {
    raf2 = window.requestAnimationFrame(() => {
      timer = window.setTimeout(task, delayMs);
    });
  });

  return () => {
    if (raf1) {
      window.cancelAnimationFrame(raf1);
    }
    if (raf2) {
      window.cancelAnimationFrame(raf2);
    }
    if (timer) {
      window.clearTimeout(timer);
    }
  };
}
