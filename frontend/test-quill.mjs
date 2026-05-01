import { JSDOM } from 'jsdom';
const dom = new JSDOM(`<body><div id="editor"></div><div id="toolbar"><select class="ql-align"></select></div></body>`);
global.document = dom.window.document;
global.window = dom.window;
import Quill from 'quill';
const quill = new Quill('#editor', { theme: 'snow', modules: { toolbar: '#toolbar' } });
console.log(document.querySelector('.ql-align').outerHTML);
