const files = import.meta.glob('./*.js', { eager: true });
const messages = {};
Object.keys(files).forEach((key) => {
  if (key === './index.js') return;
  messages[key.replace(/(\.\/|\.js)/g, '')] = files[key].default;
});
const languages = Object.keys(messages);
export { messages, languages };
