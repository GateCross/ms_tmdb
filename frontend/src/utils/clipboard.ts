/**
 * 将文本写入系统剪贴板，失败时返回 false
 */
export async function copyText(text: string): Promise<boolean> {
  if (!text) return false;
  const clipboard = window.navigator?.clipboard;
  if (!clipboard) return false;
  try {
    await clipboard.writeText(text);
    return true;
  } catch {
    return false;
  }
}
