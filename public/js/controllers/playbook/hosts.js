define([
	'app',
	'jquery',
	'services/hostgroups',
	'factories/hostgroup',
	'factories/host'
], function(app, $) {
	app.registerController('PlaybookHostsCtrl', ['$scope', '$state', 'hostgroups', 'Host', function($scope, $state, hostgroups, Host) {

		$scope.hostgroups = hostgroups;

		hostgroups.get($scope.playbook, function () {
		});

		$scope.add = function () {
			$('#addHostGroup').modal('show');
		}

		$scope.showAddHost = function (hostgroup) {
			if (hostgroup.showingAdd) {
				hostgroup.showingAdd = false;
			} else {
				hostgroup.showingAdd = true;
				hostgroup.newHost = new Host();
			}
		}

		$scope.deleteHostGroup = function (hostgroup) {
			hostgroup.delete($scope.playbook);

			hostgroups.get($scope.playbook, function () {
			});
		}

		$scope.deleteHost = function (hostgroup, host) {
			host.delete($scope.playbook, hostgroup);
			hostgroup.getHosts($scope.playbook);
		}

		$scope.addHost = function (hostgroup) {
			hostgroup.newHost.add($scope.playbook, hostgroup);
			hostgroup.getHosts($scope.playbook);

			$scope.showAddHost(hostgroup);
		}
	}]);

	app.registerController('AddHostGroupCtrl', ['$scope', 'HostGroup', 'hostgroups', function($scope, HostGroup, hostgroups) {
		$scope.hostgroup = new HostGroup();

		$scope.add = function () {
			$('#addHostGroup').modal('hide');

			$scope.hostgroup.add($scope.playbook);

			hostgroups.get($scope.playbook, function () {
			});
		}
	}]);
});