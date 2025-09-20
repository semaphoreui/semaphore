%global debug_package %{nil}
%global _missing_build_ids_terminate_build 0
%global _dwz_low_mem_die_limit 0

Name:           forge
Version:        2.8.90
Release:        1%{?dist}
Summary:        FORGE is a modern UI for Ansible, Terraform, OpenTofu, Bash and Pulumi. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

License:        MIT
URL:            https://github.com/Digital-Data-Co/forge
Source:         https://github.com/Digital-Data-Co/forge/archive/refs/tags/v2.8.90.zip

BuildRequires:  golang
BuildRequires:  nodejs
BuildRequires:  nodejs-npm
BuildRequires:  go-task
BuildRequires:  git
BuildRequires:  systemd-rpm-macros

Requires:       ansible

%description
FORGE is a modern UI for Ansible, Terraform, OpenTofu, Bash and Pulumi. It lets you easily run Ansible playbooks, get notifications about fails, control access to deployment system.

%prep
%setup -q

%build
export FORGE_VERSION="development"
export FORGE_ARCH="linux_amd64"
export FORGE_CONFIG_PATH="./etc/forge"
export APP_ROOT="./forgeui/"

if ! [[ "$PATH" =~ "$HOME/go/bin:" ]]
then
    PATH="$HOME/go/bin:$PATH"
fi
export PATH
go-task all

cat > forgeui.service <<EOF
[Unit]
Description=Forge Ansible
Documentation=https://github.com/Digital-Data-Co/forge
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecReload=/bin/kill -HUP $MAINPID
ExecStart=%{_bindir}/forge service --config=/etc/forge/config.json
SyslogIdentifier=forge
Restart=always

[Install]
WantedBy=multi-user.target

EOF

cat > forge-setup <<EOF
forge setup --config=/etc/forge/config.json
EOF

%install
mkdir -p %{buildroot}%{_sysconfdir}/forge/
mkdir -p %{buildroot}%{_bindir}
mkdir -p %{buildroot}%{_unitdir}

install -m 755 bin/forge %{buildroot}%{_bindir}/forge
install -m 755 forge-setup %{buildroot}%{_bindir}/forge-setup
install -m 755 forgeui.service %{buildroot}%{_unitdir}/forgeui.service

%files
%license LICENSE
%doc README.md CONTRIBUTING.md
%attr(755, root, root) %{_bindir}/forge
%attr(755, root, root) %{_bindir}/forge-setup
%attr(644, root,root) %{_sysconfdir}/forge/
%{_unitdir}/forgeui.service

%changelog
* Wed Jun 28 2023 Neftali Yagua
-
