app.factory('SweetAlert', [ '$rootScope', function ( $rootScope ) {
	var swal = window.swal;

	return {
		swal: function (arg1, arg2, arg3) {
			$rootScope.$evalAsync(function () {
				if (typeof(arg2) === 'function') {
					SweetAlert.swal(arg1, function (isConfirm) {
						$rootScope.$evalAsync(function () {
							arg2(isConfirm);
						});
					}, arg3);
				} else {
					SweetAlert.swal(arg1, arg2, arg3);
				}
			});
		}
	};
}]);