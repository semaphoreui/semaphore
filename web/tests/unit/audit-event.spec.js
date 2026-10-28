import './setup';
import { expect } from 'chai';
import i18n from '@/plugins/i18';
import {
  eventKind,
  eventTitle,
  reasonLabel,
  targetTypeLabel,
  actorLabel,
  actorIcon,
  metadataRows,
  filterParams,
} from '@/lib/auditEvent';

// vue-i18n 11 in composition mode: the functions take the global composer, as AuditLog.vue
// passes it; its locale is a ref.
const composer = i18n.global;

function event(fields) {
  return {
    category: 'auth', event_code: 'auth.login', action: 'authenticate', outcome: 'success', ...fields,
  };
}

describe('lib/auditEvent', () => {
  afterEach(() => {
    composer.locale.value = 'en';
  });

  it('joins category, code and action into the kind', () => {
    expect(eventKind(event())).to.equal('auth.login/authenticate');
  });

  it('titles a success and a failure differently', () => {
    expect(eventTitle(event(), composer)).to.equal('Signed in');
    expect(eventTitle(event({ outcome: 'failure' }), composer)).to.equal('Sign-in failed');
  });

  it('uses the success title for a failure without its own title', () => {
    const roleDelete = event({ category: 'iam', event_code: 'iam.role', action: 'delete' });
    expect(eventTitle({ ...roleDelete, outcome: 'failure' }, composer)).to.equal('Role deleted');
  });

  it('shows the kind code for a kind without a title', () => {
    expect(eventTitle(event({ category: 'x', event_code: 'x.y', action: 'z' }), composer)).to.equal('x.y/z');
  });

  it('keeps English titles in another locale', () => {
    composer.locale.value = 'ru';
    expect(eventTitle(event(), composer)).to.equal('Signed in');
  });

  it('labels reasons and object types, and passes unknown values through', () => {
    expect(reasonLabel('invalid_credentials', composer)).to.equal('Invalid credentials');
    expect(reasonLabel('', composer)).to.equal('');
    expect(reasonLabel('new_reason', composer)).to.equal('new_reason');
    expect(targetTypeLabel('credential', composer)).to.equal('Key');
    expect(targetTypeLabel('new_type', composer)).to.equal('new_type');
  });

  it('names each kind of actor', () => {
    expect(actorLabel({ type: 'user', name: 'alice', auth: 'session' }, composer)).to.equal('alice');
    expect(actorLabel({ type: 'user', name: 'alice', auth: 'api_token' }, composer)).to.equal('alice (API token)');
    expect(actorLabel({ type: 'runner', id: '3' }, composer)).to.equal('Runner 3');
    expect(actorLabel({ type: 'system', name: 'retention' }, composer)).to.equal('System (retention)');
    expect(actorLabel({ type: 'anonymous' }, composer)).to.equal('Anonymous');
    expect(actorLabel({ type: 'anonymous', name: 'bob' }, composer)).to.equal('bob');
  });

  it('picks an icon per actor', () => {
    expect(actorIcon({ type: 'user', auth: 'api_token' })).to.equal('mdi-api');
    expect(actorIcon({ type: 'anonymous' })).to.equal('mdi-incognito');
  });

  it('flattens metadata into rows', () => {
    expect(metadataRows({ permissions: ['run', 'manage'], auth: { method: 'ldap' }, deleted: 3 })).to.deep.equal([
      { key: 'permissions', value: 'run, manage' },
      { key: 'auth.method', value: 'ldap' },
      { key: 'deleted', value: '3' },
    ]);
    expect(metadataRows(null)).to.deep.equal([]);
  });

  it('shows a role reference by its slug', () => {
    const owner = { id: 1, slug: 'owner' };
    const runner = { id: 3, slug: 'task_runner' };
    const metadata = { old_roles: [owner], new_roles: [owner, runner], role: runner };
    expect(metadataRows(metadata)).to.deep.equal([
      { key: 'old_roles', value: 'owner' },
      { key: 'new_roles', value: 'owner, task_runner' },
      { key: 'role', value: 'task_runner' },
    ]);
  });

  it('drops empty filters', () => {
    expect(filterParams({
      user: 5, ip: '', project: null, kind: [], outcome: 'failure', from: undefined,
    })).to.deep.equal({ user: 5, outcome: 'failure' });
  });
});
