import './setup';
import { expect } from 'chai';
import { mount } from '@vue/test-utils';
import RepositoryForm from '@/components/RepositoryForm.vue';
import mockAxios from './helpers/axiosMock';

async function flush() {
  await new Promise((resolve) => { setTimeout(resolve, 0); });
  await new Promise((resolve) => { setTimeout(resolve, 0); });
}

const keys = [
  { id: 1, name: 'deploy key', type: 'ssh' },
  { id: 2, name: 'login', type: 'login_password' },
  { id: 3, name: 'anonymous', type: 'none' },
  { id: 4, name: 'pat', type: 'string' },
];

function mountForm() {
  return mount(RepositoryForm, {
    props: { projectId: 7, itemId: 'new' },
    global: { mocks: { $t: (k) => k } },
  });
}

describe('RepositoryForm.vue', () => {
  let http;

  beforeEach(() => {
    http = mockAxios();
    http.respond((c) => (c.url.endsWith('/keys') ? keys : {}));
  });

  afterEach(() => {
    http.restore();
  });

  // A string key is a token: an HTTP(S) server takes it, an SSH server does not.
  [
    { url: 'https://github.com/acme/ops.git', offered: true },
    { url: 'http://git.local/acme/ops.git', offered: true },
    { url: 'git@github.com:acme/ops.git', offered: false },
    { url: 'ssh://git@github.com/acme/ops.git', offered: false },
    { url: 'git://github.com/acme/ops.git', offered: false },
    { url: 'file:///srv/git/ops.git', offered: false },
    { url: '/srv/ops', offered: false },
    { url: '', offered: false },
  ].forEach(({ url, offered }) => {
    it(`${offered ? 'offers' : 'hides'} string keys for "${url}"`, async () => {
      const wrapper = mountForm();
      await flush();

      wrapper.vm.item.git_url = url;
      await flush();

      const names = wrapper.vm.sshKeys.map((key) => key.name);
      expect(names).to.include.members(['deploy key', 'login', 'anonymous']);
      expect(names.includes('pat')).to.equal(offered);
    });
  });
});
