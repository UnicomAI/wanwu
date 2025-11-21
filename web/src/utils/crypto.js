import CryptoJS from 'crypto-js/crypto-js'

const KEY = CryptoJS.enc.Utf8.parse('f5Su3GhNMM1rndyp');
const IV = CryptoJS.enc.Utf8.parse('sdf4ddfsFD86Vdf2');

// AESEncrypt ：String key iv  Backbase64
export function Encrypt(str, keyStr, ivStr) {
  let key = KEY;
  let iv = IV;
  if (keyStr && ivStr) {
    key = CryptoJS.enc.Utf8.parse(keyStr);
    iv = CryptoJS.enc.Utf8.parse(ivStr);
  }
  let srcs = CryptoJS.enc.Utf8.parse(str);
  var encrypted = CryptoJS.AES.encrypt(srcs, key, {
    iv: iv,
    mode: CryptoJS.mode.CBC,
    padding: CryptoJS.pad.Pkcs7,
  });
  return encrypted.toString();
}

// AES Decrypt ：String key iv  Backbase64
export function Decrypt(str, keyStr, ivStr) {
  let key = KEY;
  let iv = IV;

  if (keyStr && ivStr) {
    key = CryptoJS.enc.Utf8.parse(keyStr);
    iv = CryptoJS.enc.Utf8.parse(ivStr);
  }
  let base64 = CryptoJS.enc.Base64.parse(str);
  let src = CryptoJS.enc.Base64.stringify(base64);
  var decrypt = CryptoJS.AES.decrypt(src, key, {
    iv: iv,
    mode: CryptoJS.mode.CBC,
    padding: CryptoJS.pad.Pkcs7
  });
  var decryptedStr = decrypt.toString(CryptoJS.enc.Utf8);
  return decryptedStr.toString();
}

export function Urlencode (str) {
  str = (str + '').toString();
  return encodeURIComponent(str).replace(/!/g, '%21').replace(/'/g, '%27').replace(/\(/g, '%28').
  replace(/\)/g, '%29').replace(/\*/g, '%2A').replace(/%20/g, '+');
}


export function urlEncrypt  (data) {
    return Urlencode(Encrypt(data));
};
