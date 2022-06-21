export default function delay(milliseconds = 100) {
  return new Promise((resolve) => {
    setTimeout(resolve, milliseconds);
  });
}
