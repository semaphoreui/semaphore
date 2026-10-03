// eslint-disable-next-line import/no-extraneous-dependencies
const through = require('through2');
// eslint-disable-next-line import/no-extraneous-dependencies
const PluginError = require('plugin-error');
// eslint-disable-next-line import/no-extraneous-dependencies
const { OpenAI } = require('openai');

const PLUGIN_NAME = 'gulp-gpt-translate';

function gptTranslate(options) {
  if (!options || !options.apiKey) {
    throw new PluginError(PLUGIN_NAME, 'An OpenAI API key is required.');
  }

  if (!options.targetLanguage) {
    throw new PluginError(PLUGIN_NAME, 'A target language must be specified.');
  }

  const openai = new OpenAI();

  return through.obj((file, enc, cb) => {
    const self = this;

    if (file.isNull()) {
      return cb(null, file); // Pass along if no contents
    }

    if (file.isStream()) {
      self.emit('error', new PluginError(PLUGIN_NAME, 'Streaming not supported.'));
      return cb();
    }

    (async () => {
      try {
        const content = file.contents.toString(enc);

        const response = await openai.chat.completions.create({
          model: options.model || 'gpt-4o-mini',
          temperature: 0,
          messages: [
            {
              role: 'system',
              content: `You are a helpful assistant that translates text to ${options.targetLanguage}. `,
            },
            ...(options.messages || []).map((m) => ({ role: 'user', content: m })),
            { role: 'user', content },
          ],
        });

        // eslint-disable-next-line no-param-reassign
        file.contents = Buffer.from(`${response.choices[0].message.content}\n`, enc);

        self.push(file);
        cb();
      } catch (err) {
        self.emit('error', new PluginError(PLUGIN_NAME, err.message));
        cb(err);
      }
    })();

    return null;
  });
}

module.exports = gptTranslate;
