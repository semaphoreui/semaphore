// eslint-disable-next-line import/prefer-default-export
export function getErrorMessage(err) {
  if (err.response) {
    if (err.response.data && err.response.data.error) {
      return err.response.data.error;
    }

    if (err.message && !err.message.startsWith('Request failed with status code ')) {
      return err.message;
    }

    switch (err.response.status) {
      case 401:
        return `${err.response.status} ${err.response.statusText}`;
      default:
        return err.message;
    }
  }

  return err.message;
}
