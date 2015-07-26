define([
	'app',
	'services/users'
], function(app) {
	app.registerController('UsersCtrl', ['$scope', '$state', 'users', function($scope, $state, users) {
		users.getUsers(function () {
			$scope.users = users.users;
		});
	}]);
});
