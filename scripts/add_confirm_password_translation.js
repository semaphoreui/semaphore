const fs = require('fs');
const path = require('path');

// List of all language files
const languageFiles = [
  'ar.js', 'de.js', 'es.js', 'fr.js', 'it.js', 'ja.js',
  'ko.js', 'nl.js', 'pl.js', 'pt.js', 'pt_br.js', 'ru.js', 'zh_cn.js', 'zh_tw.js'
];

// Translation for "Confirm Password" in each language
const translations = {
  'ar.js': 'تأكيد كلمة المرور',
  'de.js': 'Passwort bestätigen',
  'es.js': 'Confirmar Contraseña',
  'fr.js': 'Confirmer le mot de passe',
  'it.js': 'Conferma Password',
  'ja.js': 'パスワード確認',
  'ko.js': '비밀번호 확인',
  'nl.js': 'Bevestig Wachtwoord',
  'pl.js': 'Potwierdź Hasło',
  'pt.js': 'Confirmar Senha',
  'pt_br.js': 'Confirmar Senha',
  'ru.js': 'Подтвердить пароль',
  'zh_cn.js': '确认密码',
  'zh_tw.js': '確認密碼'
};

const langDir = path.join(__dirname, '..', 'web', 'src', 'lang');

languageFiles.forEach(file => {
  const filePath = path.join(langDir, file);

  if (fs.existsSync(filePath)) {
    let content = fs.readFileSync(filePath, 'utf8');

    // Add the confirmPassword translation after password2
    const password2Regex = /(\s+password2:\s*['"][^'"]*['"],?\s*)/;
    const match = content.match(password2Regex);

    if (match) {
      const replacement = match[1] + `\n  confirmPassword: '${translations[file]}',`;
      content = content.replace(password2Regex, replacement);

      fs.writeFileSync(filePath, content, 'utf8');
      console.log(`Updated ${file}`);
    } else {
      console.log(`Could not find password2 in ${file}`);
    }
  } else {
    console.log(`File not found: ${file}`);
  }
});

console.log('Translation update complete!');
