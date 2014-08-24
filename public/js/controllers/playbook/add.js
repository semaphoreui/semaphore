define([
	'app',
	'factories/playbook'
], function(app) {
	app.registerController('AddPlaybookCtrl', ['$scope', 'Playbook', function($scope, Playbook) {
		$scope.playbook = new Playbook();
	}]);
});