define([
	'app',
	'services/credentials'
], function(app) {
	app.registerController('CredentialsCtrl', ['$scope', '$state', 'credentials', function($scope, $state, credentials) {
		credentials.getCredentials(function () {
			$scope.credentials = credentials.credentials;
		});
	}]);
});