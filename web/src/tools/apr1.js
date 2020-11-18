/*
 * A JavaScript implementation of the APR1 hash.
 * Version 1.0
 * Copyright (c) 2018-2019 Knivre T.
 * Distributed under the MIT license.
 * Requires md5.js from Paul Johnston's jshash: http://pajhome.org.uk/crypt/md5/
 */

import md5 from "./md5"

/**
 * Convert a raw string to an array of integers.
 * @param {string} raw_string - raw string to convert.
 * @return {array} array of integers.
 */
function rstr2array(raw_string) {
    var array = [];
    for (var i = 0; i < raw_string.length; ++i) {
        array.push(raw_string.charCodeAt(i));
    }
    return array;
}

/**
 * Convert binary data to APR's own base64 variant.
 * Note: JavaScript equivalent to crypto/apr_md5.c's to64() function.
 * @param {string} data - binary string.
 * @param {integer} size - number of 6-bit bytes to process.
 * @return {string} APR-encoded string.
 */
function bin2apr(data, size) {
    var apr_alphabet = './0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz';
    var apr_string = '';
    while (--size >= 0) {
        // Extract the lower 6 bits (63 is 00111111b), match them to the
        // adequate character in the APR alphabet above and append that
        // character to our result string:
        apr_string += apr_alphabet.charAt(63 & data);
        // Right-shift by 6 bits:
        data >>>= 6;
    }
    return apr_string;
}

/**
 * @return an 8-byte random salt value, APR-encoded.
 */
function apr1_make_salt() {
    // Allocate an array of 2 32-bit unsigned integers = 64 bits = 8 bytes:
    var byte_array = new Uint32Array(2);
    var browser_crypto = window.crypto || window.msCrypto;
    browser_crypto.getRandomValues(byte_array);
    var salt = bin2apr(byte_array[0], 4) + bin2apr(byte_array[1], 4);
    return salt;
}

/**
 * @param {string} password - password to hash.
 * @param {string} [salt] - salt value; if not provided, a random salt is generated and used.
 * @return {string} the APR1 hash for the given password and salt, using the $apr1$salt$hash format.
 */
function apr1_hash(password, salt) {
    var i, null_byte = String.fromCharCode(0);

    // Define the Magic String prefix that identifies a password as being
    // hashed using APR:
    var apr1_id = '$apr1$';

    // Generate a random salt if none was provided:
    if (!salt) {
        salt = apr1_make_salt();
    }

    // The APR magic itself, for a total of 1002 MD5 rounds:
    var context = password + apr1_id + salt;
    var psp_hash = md5.rstr_md5(password + salt + password);
    for (var i = password.length; i > 0; i -= 16) {
        context += psp_hash.substr(0, i > 16 ? 16 : i);
    }
    for (i = password.length; i !== 0; i >>= 1) {
        context += (1 & i) ? null_byte : password.charAt(0);
    }
    let hash = md5.rstr_md5(context);
    for (i = 0; i < 1000; ++i) {
        context = ((1 & i) ? password : hash.substr(0, 16));
        if (i % 3) context += salt;
        if (i % 7) context += password;
        context += ((1 & i) ? hash.substr(0, 16) : password);
        hash = md5.rstr_md5(context);
    }

    // APR-encode the resulting 128-bit hash:
    hash = rstr2array(hash);
    var apr_hash;
    // Pass 3 8-bit bytes (i.e. 24 bits) to bin2apr and instruct it to
    // process 4 6-bit bytes (i.e. 24 bits):
    apr_hash = bin2apr(hash[0] << 16 | hash[6] << 8 | hash[12], 4); // total:  24 bits
    apr_hash += bin2apr(hash[1] << 16 | hash[7] << 8 | hash[13], 4); // total:  48 bits
    apr_hash += bin2apr(hash[2] << 16 | hash[8] << 8 | hash[14], 4); // total:  72 bits
    apr_hash += bin2apr(hash[3] << 16 | hash[9] << 8 | hash[15], 4); // total:  96 bits
    apr_hash += bin2apr(hash[4] << 16 | hash[10] << 8 | hash[5], 4); // total: 120 bits
    apr_hash += bin2apr(hash[11], 2);                                  // total: 128 bits
    return apr1_id + salt + '$' + apr_hash;
}

/**
 * @param {string} string - string to check.
 * @return {bool} true if the given string appears to be an APR1 hash, false otherwise.
 */
function apr1_is_hash(string) {
    return string.search(/^\$apr1\$[./0-9A-Za-z]{8}\$[./0-9A-Za-z]{22}$/) >= 0;
}

export default {
    hash: apr1_hash
}
