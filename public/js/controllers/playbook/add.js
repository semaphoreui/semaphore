define([
	'app',
	'factories/playbook',
	'services/playbooks',
	'services/identities'
], function(app) {
	app.registerController('AddPlaybookCtrl', ['$scope', 'Playbook', 'playbooks', '$state', 'identities', function($scope, Playbook, playbooks, $state, identities) {
		$scope.playbook = new Playbook();

		identities.getIdentities(function () {
			$scope.identities = identities.identities;
			if (!$scope.$$phase) {
				$scope.$digest();
			}
		});

		$scope.add = function () {
			$scope.playbook.add()
			.success(function (data) {
				playbooks.getPlaybooks(function () {
					$state.transitionTo('playbook.tasks', {
						playbook_id: data._id
					});
				});
			})
			.error(function (data) {

			})
		}
	}]);
});