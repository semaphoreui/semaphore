define([
	'app',
	'services/playbooks'
], function(app) {
	app.registerController('PlaybooksCtrl', ['$scope', 'playbooks', function($scope, playbooks) {
		$scope.playbooks = playbooks.playbooks;
	}]);
});