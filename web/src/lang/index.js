const files = require.context('.', false, /\.js$/);
const messages = {};
files.keys().forEach((key) => {
  if (key === './index.js') return;
  messages[key.replace(/(\.\/|\.js)/g, '')] = files(key).default;
});
const languages = Object.keys(messages);
export { messages, languages };
