define([
	'app',
	'factories/playbook',
	'services/playbooks',
	'services/credentials'
], function(app) {
	app.registerController('AddPlaybookCtrl', ['$scope', 'Playbook', 'playbooks', '$state', 'credentials', function($scope, Playbook, playbooks, $state, credentials) {
		$scope.playbook = new Playbook();
		
		credentials.getCredentials(function () {
			$scope.credentials = credentials.credentials;
		});

		$scope.add = function () {
			$scope.playbook.add()
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