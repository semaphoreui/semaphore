%global debug_package %{nil}
%global _missing_build_ids_terminate_build 0
%global _dwz_low_mem_die_limit 0

Name:           semaphore
Version:        2.8.90
Release:        1%{?dist}
Summary:        Semaphore UI is a modern UI for Ansible, Terraform, OpenTofu, Bash and Pulumi. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

License:        MIT
URL:            https://github.com/ansible-semaphore/semaphore
Source:         https://github.com/ansible-semaphore/semaphore/archive/refs/tags/v2.8.90.zip

BuildRequires:  golang
BuildRequires:  nodejs
BuildRequires:  nodejs-npm
BuildRequires:  go-task
BuildRequires:  git
BuildRequires:  systemd-rpm-macros

Requires:       ansible

%description
Semaphore UI is a modern UI for Ansible, Terraform, OpenTofu, Bash and Pulumi. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

%prep
%setup -q

%build
export SEMAPHORE_VERSION="development"
export SEMAPHORE_ARCH="linux_amd64"
export SEMAPHORE_CONFIG_PATH="./etc/semaphore"
export APP_ROOT="./ansible-semaphore/"

if ! [[ "$PATH" =~ "$HOME/go/bin:" ]]
then
    PATH="$HOME/go/bin:$PATH"
fi
export PATH
go-task all

cat > ansible-semaphore.service <<EOF
[Unit]
Description=Semaphore Ansible
Documentation=https://github.com/ansible-semaphore/semaphore
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=%{_bindir}/semaphore service --config=/etc/semaphore/config.json
SyslogIdentifier=semaphore
Restart=always

[Install]
WantedBy=multi-user.target

EOF

cat > semaphore-setup <<EOF
semaphore setup --config=/etc/semaphore/config.json
EOF

%install
mkdir -p %{buildroot}%{_sysconfdir}/semaphore/
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}

install -m 755 bin/semaphore %{buildroot}%{_bindir}/semaphore
install -m 755 semaphore-setup %{buildroot}%{_bindir}/semaphore-setup
install -m 755 ansible-semaphore.service %{buildroot}%{_unitdir}/ansible-semaphore.service

%files
%license LICENSE
%doc README.md CONTRIBUTING.md
%attr(755, root, root) %{_bindir}/semaphore
%attr(755, root, root) %{_bindir}/semaphore-setup
%attr(644, root,root) %{_sysconfdir}/semaphore/
%{_unitdir}/ansible-semaphore.service

%changelog
* Wed Jun 28 2023 Neftali Yagua
-
