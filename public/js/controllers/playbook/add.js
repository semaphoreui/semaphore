define([
	'app',
	'factories/playbook',
	'services/playbooks'
], function(app) {
	app.registerController('AddPlaybookCtrl', ['$scope', 'Playbook', 'playbooks', '$state', function($scope, Playbook, playbooks, $state) {
		$scope.playbook = new Playbook();

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