define([
	'services/users'
], function () {
	app.registerController('UsersCtrl', ['$scope', '$state', 'users', function($scope, $state, users) {
		users.getUsers(function () {
			$scope.users = users.users;
		});
	}]);
});
