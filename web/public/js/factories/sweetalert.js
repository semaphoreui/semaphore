app.factory('SweetAlert', [ '$rootScope', function ( $rootScope ) {
	return {
		swal: function () {
			var args = arguments;
			return new Promise(function (resolve, reject) {
				$rootScope.$evalAsync(function () {
					window.swal.apply(null, args).then(resolve);
				});
			});
		}
	};
}]);