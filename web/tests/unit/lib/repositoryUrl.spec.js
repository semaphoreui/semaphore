import { expect } from 'chai';
import { getRepositoryUrlType, isLocalRepositoryPath } from '@/lib/repositoryUrl';

describe('getRepositoryUrlType', () => {
  it('returns null for null', () => {
    expect(getRepositoryUrlType(null)).to.equal(null);
  });

  it('returns null for undefined', () => {
    expect(getRepositoryUrlType(undefined)).to.equal(null);
  });

  it('returns null for empty string', () => {
    expect(getRepositoryUrlType('')).to.equal(null);
  });

  it('returns local for Unix absolute path', () => {
    expect(getRepositoryUrlType('/abs/path')).to.equal('local');
    expect(getRepositoryUrlType('/home/user/repo')).to.equal('local');
  });

  it('returns local for Windows drive path with backslash', () => {
    expect(getRepositoryUrlType('D:\\repo')).to.equal('local');
    expect(getRepositoryUrlType('C:\\Users\\user\\repo')).to.equal('local');
  });

  it('returns local for Windows drive path with forward slash', () => {
    expect(getRepositoryUrlType('D:/repo')).to.equal('local');
    expect(getRepositoryUrlType('C:/Users/user/repo')).to.equal('local');
  });

  it('returns local for Windows drive letter only', () => {
    expect(getRepositoryUrlType('D:')).to.equal('local');
  });

  it('returns local for UNC path', () => {
    expect(getRepositoryUrlType('\\\\server\\share')).to.equal('local');
    expect(getRepositoryUrlType('\\\\server\\share\\repo')).to.equal('local');
  });

  it('returns ssh for git@host:repo style', () => {
    expect(getRepositoryUrlType('git@github.com:user/repo')).to.equal('ssh');
    expect(getRepositoryUrlType('git@host:repo')).to.equal('ssh');
  });

  it('returns https for https:// URLs', () => {
    expect(getRepositoryUrlType('https://github.com/user/repo')).to.equal('https');
  });

  it('returns http for http:// URLs', () => {
    expect(getRepositoryUrlType('http://github.com/user/repo')).to.equal('http');
  });

  it('returns git for git:// URLs', () => {
    expect(getRepositoryUrlType('git://github.com/user/repo')).to.equal('git');
  });

  it('returns file for file:// URLs', () => {
    expect(getRepositoryUrlType('file:///home/user/repo')).to.equal('file');
  });

  it('returns ssh for ssh:// URLs', () => {
    expect(getRepositoryUrlType('ssh://git@github.com/user/repo')).to.equal('ssh');
  });

  it('returns null for unknown URL schemes', () => {
    expect(getRepositoryUrlType('ftp://example.com/repo')).to.equal(null);
    expect(getRepositoryUrlType('svn://example.com/repo')).to.equal(null);
  });
});

describe('isLocalRepositoryPath', () => {
  it('returns true for Unix absolute path', () => {
    expect(isLocalRepositoryPath('/abs/path')).to.equal(true);
  });

  it('returns true for Windows drive path', () => {
    expect(isLocalRepositoryPath('D:\\repo')).to.equal(true);
    expect(isLocalRepositoryPath('D:/repo')).to.equal(true);
  });

  it('returns true for UNC path', () => {
    expect(isLocalRepositoryPath('\\\\server\\share')).to.equal(true);
  });

  it('returns false for https URL', () => {
    expect(isLocalRepositoryPath('https://github.com/user/repo')).to.equal(false);
  });

  it('returns false for SSH URL', () => {
    expect(isLocalRepositoryPath('git@github.com:user/repo')).to.equal(false);
  });

  it('returns false for null', () => {
    expect(isLocalRepositoryPath(null)).to.equal(false);
  });
});
