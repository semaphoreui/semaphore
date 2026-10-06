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

function event(fields) {
  return {
    category: 'auth', event_code: 'login', action: 'authenticate', outcome: 'success', ...fields,
  };
}

describe('lib/auditEvent', () => {
  afterEach(() => {
    i18n.locale = 'en';
  });

  it('joins category, code and action into the kind', () => {
    expect(eventKind(event())).to.equal('auth.login/authenticate');
  });

  it('titles a success and a failure differently', () => {
    expect(eventTitle(event(), i18n)).to.equal('Signed in');
    expect(eventTitle(event({ outcome: 'failure' }), i18n)).to.equal('Sign-in failed');
  });

  it('uses the success title for a failure without its own title', () => {
    const roleDelete = event({ category: 'iam', event_code: 'role', action: 'delete' });
    expect(eventTitle({ ...roleDelete, outcome: 'failure' }, i18n)).to.equal('Role deleted');
  });

  it('shows the kind code for a kind without a title', () => {
    expect(eventTitle(event({ category: 'x', event_code: 'y', action: 'z' }), i18n)).to.equal('x.y/z');
  });

  it('keeps English titles in another locale', () => {
    i18n.locale = 'ru';
    expect(eventTitle(event(), i18n)).to.equal('Signed in');
  });

  it('labels reasons and object types, and passes unknown values through', () => {
    expect(reasonLabel('invalid_credentials', i18n)).to.equal('Invalid credentials');
    expect(reasonLabel('', i18n)).to.equal('');
    expect(reasonLabel('new_reason', i18n)).to.equal('new_reason');
    expect(targetTypeLabel('credential', i18n)).to.equal('Key');
    expect(targetTypeLabel('new_type', i18n)).to.equal('new_type');
  });

  it('names each kind of actor', () => {
    expect(actorLabel({ type: 'user', name: 'alice', auth: 'session' }, i18n)).to.equal('alice');
    expect(actorLabel({ type: 'user', name: 'alice', auth: 'api_token' }, i18n)).to.equal('alice (API token)');
    expect(actorLabel({ type: 'runner', id: '3' }, i18n)).to.equal('Runner 3');
    expect(actorLabel({ type: 'system', name: 'retention' }, i18n)).to.equal('System (retention)');
    expect(actorLabel({ type: 'anonymous' }, i18n)).to.equal('Anonymous');
    expect(actorLabel({ type: 'anonymous', name: 'bob' }, i18n)).to.equal('bob');
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

  it('drops empty filters', () => {
    expect(filterParams({
      user: 5, ip: '', project: null, kind: [], outcome: 'failure', from: undefined,
    })).to.deep.equal({ user: 5, outcome: 'failure' });
  });
});
