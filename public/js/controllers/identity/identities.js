define([
	'app',
	'services/identities'
], function(app) {
	app.registerController('IdentitiesCtrl', ['$scope', '$state', 'identities', function($scope, $state, identities) {
		identities.getIdentities(function () {
			$scope.identities = identities.identities;
		});
	}]);
});