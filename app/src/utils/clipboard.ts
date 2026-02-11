// copyToClipboard copies the text parameter to the users clipboard.
//
// The function will return void if successful or throw an error if not.
export async function copyToClipboard(text: string): Promise<void> {
  if (navigator.clipboard) {
    try {
      await navigator.clipboard.writeText(text);
      return;
    } catch (error) {
      console.error(error);
    }
  }

  const err = new Error();
  err.message = `Browser does not support copy functionality. Manually copy the text:\n\n ${Error}`;
  throw err;
}
