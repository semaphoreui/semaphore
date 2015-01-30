define([
	'app',
	'factories/playbook',
	'services/playbooks',
	'services/identities'
], function(app) {
	app.registerController('EditPlaybookCtrl', ['$scope', 'playbook', 'playbooks', '$state', 'identities', function($scope, playbook, playbooks, $state, identities) {
		$scope.playbook = playbook;

		identities.getIdentities(function () {
			$scope.identities = identities.identities;
			if (!$scope.$$phase) {
				$scope.$digest();
			}
		});

		$scope.add = function () {
			$scope.playbook.save()
			.success(function (data) {
				playbooks.getPlaybooks(function () {
					$state.transitionTo('playbook.view', {
						playbook_id: data._id
					});
				});
			})
			.error(function (data) {
			})
		}
	}]);
});