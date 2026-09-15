import { expect } from 'chai';

describe('Keys.vue', () => {
  let Keys;

  before(() => {
    Object.defineProperty(global, 'localStorage', {
      configurable: true,
      value: {
        getItem: () => null,
      },
    });
    Object.defineProperty(global, 'navigator', {
      configurable: true,
      value: {
        language: 'en-US',
      },
    });
    // eslint-disable-next-line global-require
    Keys = require('@/views/project/Keys.vue').default;
  });

  it('does not show the generated public key dialog for manually entered keys on create', async () => {
    const ctx = {
      items: [],
      createdPublicKey: 'stale',
      createdPublicKeyDialog: false,
      loadItems: async () => {},
      extractPublicKey: () => 'ssh-rsa AAAA',
    };

    await Keys.methods.loadItemsAndShowPublicKey.call(ctx, {
      action: 'new',
      item: {
        id: 1,
        generate_ssh_key: false,
        plain: '{"public_key":"ssh-rsa AAAA"}',
      },
    });

    expect(ctx.createdPublicKey).to.equal('');
    expect(ctx.createdPublicKeyDialog).to.equal(false);
  });

  it('shows the generated public key dialog when a key was generated on create', async () => {
    const item = {
      id: 1,
      generate_ssh_key: true,
      plain: '{"public_key":"ssh-rsa AAAA"}',
    };
    const ctx = {
      items: [item],
      createdPublicKey: '',
      createdPublicKeyDialog: false,
      loadItems: async () => {},
      extractPublicKey: Keys.methods.extractPublicKey,
    };

    await Keys.methods.loadItemsAndShowPublicKey.call(ctx, {
      action: 'new',
      item,
    });

    expect(ctx.createdPublicKey).to.equal('ssh-rsa AAAA');
    expect(ctx.createdPublicKeyDialog).to.equal(true);
  });
});
