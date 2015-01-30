define([
	'app',
	'services/playbooks'
], function(app) {
	app.registerController('PlaybookCtrl', ['$scope', '$state', 'playbooks', 'playbook', function($scope, $state, playbooks, playbook) {
		$scope.playbook = playbook;

		$scope.delete = function () {
			$scope.playbook.delete();

			playbooks.getPlaybooks(function () {
				$state.transitionTo('homepage');
			})
		}
	}]);
});