define([
	'app'
], function(app) {
	app.registerController('PlaybookCtrl', ['$scope', function($scope) {
		console.log($scope.playbook);
	}]);
});