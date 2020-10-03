// eslint-disable-next-line import/prefer-default-export
export function getErrorMessage(err) {
  if (err.response && err.response.data) {
    return err.response.data.error || err.message;
  }
  return err.message;
}
