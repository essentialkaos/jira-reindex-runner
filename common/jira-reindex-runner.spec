################################################################################

%global crc_check pushd ../SOURCES ; sha512sum -c %{SOURCE100} ; popd

################################################################################

%define debug_package     %{nil}

################################################################################

Summary:        Application for periodical running Jira re-index process
Name:           jira-reindex-runner
Version:        0.0.5
Release:        0%{?dist}
Group:          Applications/System
License:        Apache License, Version 2.0
URL:            https://kaos.sh/jira-reindex-runner

Source0:        https://source.kaos.st/%{name}/%{name}-%{version}.tar.bz2

Source100:      checksum.sha512

BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildRequires:  golang >= 1.19

Provides:       %{name} = %{version}-%{release}

################################################################################

%description
Application for periodical running Jira re-index process.

################################################################################

%prep
%{crc_check}

%setup -q

%build
if [[ ! -d "%{name}/vendor" ]] ; then
  echo "This package requires vendored dependencies"
  exit 1
fi

pushd %{name}
  go build %{name}.go
  cp LICENSE ..
popd

%install
rm -rf %{buildroot}

install -dm 755 %{buildroot}%{_bindir}
install -dm 755 %{buildroot}%{_sysconfdir}
install -dm 755 %{buildroot}%{_sysconfdir}/cron.d
install -dm 755 %{buildroot}%{_sysconfdir}/logrotate.d
install -dm 755 %{buildroot}%{_logdir}/%{name}
install -dm 755 %{buildroot}%{_mandir}/man1

pushd %{name}
  install -pm 755 %{name} \
                  %{buildroot}%{_bindir}/

  install -pm 600 common/%{name}.knf \
                  %{buildroot}%{_sysconfdir}/

  install -pm 644 common/%{name}.logrotate \
                  %{buildroot}%{_sysconfdir}/logrotate.d/%{name}

  install -pm 644 common/%{name}.cron \
                  %{buildroot}%{_sysconfdir}/cron.d/%{name}

  ./%{name} --generate-man > %{buildroot}%{_mandir}/man1/%{name}.1
popd

%post
if [[ -d %{_sysconfdir}/bash_completion.d ]] ; then
  %{name} --completion=bash 1> %{_sysconfdir}/bash_completion.d/%{name} 2>/dev/null
fi

if [[ -d %{_datarootdir}/fish/vendor_completions.d ]] ; then
  %{name} --completion=fish 1> %{_datarootdir}/fish/vendor_completions.d/%{name}.fish 2>/dev/null
fi

if [[ -d %{_datadir}/zsh/site-functions ]] ; then
  %{name} --completion=zsh 1> %{_datadir}/zsh/site-functions/_%{name} 2>/dev/null
fi

%postun
if [[ $1 == 0 ]] ; then
  if [[ -f %{_sysconfdir}/bash_completion.d/%{name} ]] ; then
    rm -f %{_sysconfdir}/bash_completion.d/%{name} &>/dev/null || :
  fi

  if [[ -f %{_datarootdir}/fish/vendor_completions.d/%{name}.fish ]] ; then
    rm -f %{_datarootdir}/fish/vendor_completions.d/%{name}.fish &>/dev/null || :
  fi

  if [[ -f %{_datadir}/zsh/site-functions/_%{name} ]] ; then
    rm -f %{_datadir}/zsh/site-functions/_%{name} &>/dev/null || :
  fi
fi

%clean
rm -rf %{buildroot}

################################################################################

%files
%defattr(-,root,root,-)
%doc LICENSE
%dir %{_logdir}/%{name}
%config(noreplace) %{_sysconfdir}/%{name}.knf
%config(noreplace) %{_sysconfdir}/cron.d/%{name}
%config(noreplace) %{_sysconfdir}/logrotate.d/%{name}
%{_mandir}/man1/%{name}.1.*
%{_bindir}/%{name}

################################################################################

%changelog
* Fri Jul 14 2023 Anton Novojilov <andy@essentialkaos.com> - 0.0.5-0
- Dependencies update

* Wed Mar 30 2022 Anton Novojilov <andy@essentialkaos.com> - 0.0.4-0
- Removed pkg.re usage
- Added module info
- Added Dependabot configuration

* Thu Jul 29 2021 Anton Novojilov <andy@essentialkaos.com> - 0.0.3-0
- Fixed bug with checking re-index progress

* Wed Jul 28 2021 Anton Novojilov <andy@essentialkaos.com> - 0.0.2-0
- Fixed bug with handling main enable switch in configuration file

* Wed Jul 28 2021 Anton Novojilov <andy@essentialkaos.com> - 0.0.1-1
- Fixed permissions for configuration file

* Sat Jul 17 2021 Anton Novojilov <andy@essentialkaos.com> - 0.0.1-0
- Initial build for kaos-repo
