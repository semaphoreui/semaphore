define([
	'app',
	'services/playbooks'
], function(app) {
	app.registerController('PlaybookCtrl', ['$scope', '$state', 'playbooks', function($scope, $state, playbooks) {
		console.log($scope.playbook);

		$scope.delete = function () {
			$scope.playbook.delete();

			playbooks.getPlaybooks(function () {
				$state.transitionTo('homepage');
			})
		}
	}]);
});